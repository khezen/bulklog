package server

import (
	"io/ioutil"
	"net/http"
	"strings"

	"github.com/bulklog/bulklog/collection"
)

// POST /v1/{collection}/{schema}
func (s *Server) handleCollect(w http.ResponseWriter, r *http.Request) {
	urlSplit := strings.Split(strings.Trim(strings.ToLower(r.URL.Path), "/"), "/")
	if len(urlSplit) != 3 {
		s.serveError(w, r, ErrPathNotFound)
		return
	}
	if r.Method != http.MethodPost {
		s.serveError(w, r, ErrWrongMethod)
	}
	collectionName := collection.Name(urlSplit[1])
	schemaName := collection.SchemaName(urlSplit[2])
	docBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.serveError(w, r, err)
		return
	}
	err = s.engine.Collect(collectionName, schemaName, docBytes)
	if err != nil {
		s.serveError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GET /v1/liveness
func (s *Server) handleLiveness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

// GET /v1/readiness
func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}
