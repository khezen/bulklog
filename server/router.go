package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/bulklog/bulklog/collection"
	"github.com/bulklog/bulklog/log"
)

// ListenAndServe - Blocks the current goroutine, opens an HTTP port and serves the web REST requests
func (s *Server) ListenAndServe() {
	http.HandleFunc("/liveness", s.handleLiveness)
	http.HandleFunc("/readiness", s.handleReadiness)
	http.HandleFunc("/v1/", s.handleCollection)
	endpoint := fmt.Sprintf(":%d", s.port)
	log.Out().Printf("opening bulklog at %v\n", endpoint)
	s.quit <- http.ListenAndServe(endpoint, nil)
}

func (s *Server) handleCollection(w http.ResponseWriter, r *http.Request) {
	urlSplit := strings.Split(strings.Trim(strings.ToLower(r.URL.Path), "/"), "/")
	urlSplitLen := len(urlSplit)
	if urlSplitLen < 3 {
		s.serveError(w, r, ErrPathNotFound)
		return
	}
	collectionName := collection.Name(collection.Name(urlSplit[1]))
	schemaName := collection.SchemaName(collection.SchemaName(urlSplit[2]))
	switch urlSplitLen {
	case 3:
		switch r.Method {
		case http.MethodPost:
			s.handleCollect(w, r, collectionName, schemaName)
			return
		default:
			s.serveError(w, r, ErrWrongMethod)
			return
		}
	case 4:
		if urlSplit[3] != "batch" {
			s.serveError(w, r, ErrPathNotFound)
			return
		}
		switch r.Method {
		case http.MethodPost:
			s.handleCollectBatch(w, r, collectionName, schemaName)
			return
		default:
			s.serveError(w, r, ErrWrongMethod)
			return
		}
	default:
		s.serveError(w, r, ErrPathNotFound)
		return
	}
}
