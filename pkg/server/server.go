package server

import (
	"github.com/khezen/bulklog/pkg/config"
	"github.com/khezen/bulklog/pkg/engine"
)

const defaultPort = 5017

// Server - Contains data required for serving web REST requests
type Server struct {
	port   int
	engine engine.Engine
	quit   chan error
}

// New - Create new service for serving web REST requests
func New(cfg *config.Config, quit chan error) (*Server, error) {
	e, err := engine.New(cfg)
	if err != nil {
		return nil, err
	}
	port := cfg.Port
	if port == 0 {
		port = defaultPort
	}
	srv := Server{
		port,
		e,
		quit,
	}
	return &srv, nil
}
