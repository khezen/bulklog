package server

import (
	"fmt"
	"net/http"

	config "github.com/khezen/bulklog/config"
	"github.com/khezen/bulklog/engine"
)

const defaultPort = 5000

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

// ListenAndServe - Blocks the current goroutine, opens an HTTP port and serves the web REST requests
func (s *Server) ListenAndServe() {
	http.HandleFunc("/v1/liveness", s.handleLiveness)
	http.HandleFunc("/v1/readiness", s.handleReadiness)
	http.HandleFunc("/v1/", s.handleCollect)
	endpoint := fmt.Sprintf(":%d", s.port)
	fmt.Printf("opening bulklog at %v\n", endpoint)
	s.quit <- http.ListenAndServe(endpoint, nil)
}
