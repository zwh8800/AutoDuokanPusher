package conf

import (
	"flag"

	"github.com/BurntSushi/toml"
)

type config struct {
	Sender struct {
		Username string
		Password string
		Host     string
		Port     int
		SSL      bool
		Name     string
		Address  string
	}

	Receiver struct {
		Name    string
		Address string
	}

	Cc struct {
		Name    string
		Address string
	}

	Watcher struct {
		Path string
	}
}

var Conf config

func init() {
	configPath := flag.String("config", "duokan.toml", "specify a config file")
	flag.Parse()
	if _, err := toml.DecodeFile(*configPath, &Conf); err != nil {
		panic(err)
	}
}
