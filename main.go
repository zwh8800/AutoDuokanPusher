package main

import (
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/zwh8800/duokan-pusher/conf"
	"gopkg.in/gomail.v2"
)

func main() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()
	if err := watcher.Add(conf.Conf.Watcher.Path); err != nil {
		panic(err)
	}
	log.Println("watching")
	for e := range watcher.Events {
		log.Println("event:", e)
		if e.Op&fsnotify.Create == fsnotify.Create {
			go queueToSend(e.Name)
		} else if e.Op&fsnotify.Write == fsnotify.Write {
			go waitOneMinute(e.Name)
		}
	}
}

var timers map[string]*time.Timer = make(map[string]*time.Timer)
var timersLock sync.RWMutex

func queueToSend(path string) {
	timersLock.RLock()
	if _, ok := timers[path]; ok {
		timersLock.RUnlock()
		return
	}
	timersLock.RUnlock()

	timer := time.NewTimer(1 * time.Minute)
	timersLock.Lock()
	timers[path] = timer
	timersLock.Unlock()

	<-timer.C
	timersLock.Lock()
	delete(timers, path)
	timersLock.Unlock()
	timer.Stop()

	sendEmail(path)
}

func waitOneMinute(path string) {
	timersLock.RLock()
	defer timersLock.RUnlock()
	timer, ok := timers[path]
	if !ok {
		return
	}
	timer.Reset(1 * time.Minute)
}

func sendEmail(path string) {
	log.Println("sending file:", path)

	m := gomail.NewMessage()
	m.SetAddressHeader("From", conf.Conf.Sender.Address, conf.Conf.Sender.Name)
	m.SetAddressHeader("To", conf.Conf.Receiver.Address, conf.Conf.Receiver.Name)
	if conf.Conf.Cc.Address != "" {
		m.SetAddressHeader("Cc", conf.Conf.Cc.Address, conf.Conf.Cc.Name)
	}
	m.Attach(path)

	d := gomail.NewDialer(conf.Conf.Sender.Host, conf.Conf.Sender.Port, conf.Conf.Sender.Username, conf.Conf.Sender.Password)
	d.SSL = conf.Conf.Sender.SSL
	if err := d.DialAndSend(m); err != nil {
		log.Println("Error: ", err)
	}
	log.Println("send", path, "ok")
}
