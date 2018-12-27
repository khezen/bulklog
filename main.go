package main

import (
	"github.com/khezen/bulklog/config"
	"github.com/khezen/bulklog/server"
)

func main() {
	quit := make(chan error)
	var err error
	cfg, err := config.Get()
	if err != nil {
		panic(err)
	}
	server, err := server.New(cfg, quit)
	if err != nil {
		panic(err)
	}
	go server.ListenAndServe()
	err = <-quit
	panic(err)
}
