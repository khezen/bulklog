package server

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/khezen/bulklog/internal/collection"
)

// POST /v1/{collection}/{schema}
func (s *Server) handleCollect(w http.ResponseWriter, r *http.Request, collectionName collection.Name, schemaName collection.SchemaName) {
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

// POST /v1/{collection}/{schemaName}/batch
func (s *Server) handleCollectBatch(w http.ResponseWriter, r *http.Request, collectionName collection.Name, schemaName collection.SchemaName) {
	docsBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.serveError(w, r, err)
		return
	}
	buf := bytes.NewBuffer(docsBytes)
	length := bytes.Count(docsBytes, []byte("\n"))
	docBytesSlice := make([][]byte, 0, length)
	for {
		docBytes, err := buf.ReadBytes('\n')
		if len(docBytes) == 0 {
			break
		}
		docBytesSlice = append(docBytesSlice, docBytes)
		if err != nil {
			break
		}
	}
	err = s.engine.CollectBatch(collectionName, schemaName, docBytesSlice...)
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
