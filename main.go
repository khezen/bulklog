package main

import (
	"github.com/khezen/espipe/configuration"
	"github.com/khezen/espipe/service"
)

var (
	configFile = "/etc/espipe/config.json"
)

func main() {
	quit := make(chan error)

	config, err := configuration.LoadConfig(configFile)
	if err != nil {
		config, err = configuration.LoadConfig("config.json")
		if err != nil {
			panic(err)
		}
	}

	service, err := service.New(config, quit)
	if err != nil {
		panic(err)
	}

	go service.ListenAndServe()
	err = <-quit
	panic(err)
}
