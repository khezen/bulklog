package server

import (
	"fmt"
	"net/http"

	config "github.com/khezen/espipe/config"
	"github.com/khezen/espipe/engine"
)

const endpoint = ":5000"

// Server - Contains data required for serving web REST requests
type Server struct {
	engine engine.Engine
	quit   chan error
}

// New - Create new service for serving web REST requests
func New(cfg config.Config, quit chan error) (*Server, error) {
	e, err := engine.New(cfg)
	if err != nil {
		return nil, err
	}
	srv := Server{
		e,
		quit,
	}
	return &srv, nil
}

// ListenAndServe - Blocks the current goroutine, opens an HTTP port and serves the web REST requests
func (s *Server) ListenAndServe() {
	http.HandleFunc("/espipe/v1/health/", s.handleHealthCheck)
	http.HandleFunc("/espipe/v1/", s.handleCollect)
	fmt.Printf("opening espipe at %v\n", endpoint)
	s.quit <- http.ListenAndServe(endpoint, nil)
}