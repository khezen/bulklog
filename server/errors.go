package server

import (
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/khezen/bulklog/engine"
)

var (
	// ErrPathNotFound - 404
	ErrPathNotFound = errors.New("ErrPathNotFound - The request path is not supported")
	// ErrWrongMethod - 405
	ErrWrongMethod = errors.New("ErrWrongMethod - The request http method does not match expectation")
)

// HTTPStatusCode -
func HTTPStatusCode(err error) int {
	switch err {
	case ErrPathNotFound, engine.ErrNotFound:
		return 404
	case ErrWrongMethod:
		return 405
	default:
		return 500
	}
}

func (s *Server) serveError(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Printf("%v", err)
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	statusCode := HTTPStatusCode(err)
	w.WriteHeader(statusCode)
	io.WriteString(w, err.Error())
}
