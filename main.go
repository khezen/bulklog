package main

import (
	"github.com/khezen/bulklog/config"
	"github.com/khezen/bulklog/server"
)

var (
	configFile = "/etc/bulklog/config.yaml"
)

func main() {
	quit := make(chan error)
	var err error
	cfg, err := config.Get()
	if cfg == nil {
		config.Set("config.yaml")
		cfg, err = config.Get()
	}
	server, err := server.New(cfg, quit)
	if err != nil {
		panic(err)
	}
	go server.ListenAndServe()
	err = <-quit
	panic(err)
}
