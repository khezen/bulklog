package service

import (
	"fmt"
	"io"
	"net/http"

	"github.com/khezen/espipe/errors"
)

func (s *Service) serveError(w http.ResponseWriter, r *http.Request, err error) {
	w.Header().Set("Connection", "close")
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	switch err {
	case errors.ErrPathNotFound:
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, err.Error())
		return
	case errors.ErrWrongMethod:
		w.WriteHeader(http.StatusMethodNotAllowed)
		io.WriteString(w, err.Error())
		return
	}
	fmt.Printf("%v", err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	io.WriteString(w, err.Error())
}
