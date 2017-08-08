package main

import (
	"fmt"
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
			fatal(err)
		}
	}

	service, err := service.New(config, quit)
	if err != nil {
		fatal(err)
	}

	go service.ListenAndServe()
	err = <-quit
	fatal(err)
}

func fatal(err error) {
	fmt.Println(err)
	panic(err) // Nothing more we can do ... just panic ... result in returning error code
}
