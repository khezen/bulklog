package service

import (
	"fmt"
	"github.com/khezen/espipe/errors"
	"io"
	"net/http"
)

func (s *Service) serveError(w http.ResponseWriter, r *http.Request, err error) {

	switch err {
	case errors.ErrPathNotFound:
		w.WriteHeader(http.StatusNotFound)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, err.Error())
		return
	case errors.ErrWrongMethod:
		w.WriteHeader(http.StatusMethodNotAllowed)
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		io.WriteString(w, err.Error())
		return
	}
	fmt.Printf("%v", err.Error())
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	io.WriteString(w, err.Error())
}
