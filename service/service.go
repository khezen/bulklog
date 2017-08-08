package service

import (
	"fmt"
	configuration "github.com/khezen/espipe/configuration"
	dispatcher "github.com/khezen/espipe/dispatcher"
	errors "github.com/khezen/espipe/errors"
	model "github.com/khezen/espipe/model"
	"github.com/khezen/espipe/uuid"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

const httpHeaderXRequestID = "REQUEST-ID"

// Service - Contains data required for serving web REST requests
type Service struct {
	config             configuration.Configuration
	Dispatcher         *dispatcher.Dispatcher
	availableTemplates map[configuration.TemplateName]configuration.Template
	availableResources map[configuration.TemplateName]map[model.DocumentType]bool
	quit               chan error
}

// New - Create new service for serving web REST requests
func New(config configuration.Configuration, quit chan error) (*Service, error) {
	d, err := dispatcher.NewDispatcher(&config)

	if err != nil {
		return nil, err
	}
	availableTemplates := make(map[configuration.TemplateName]configuration.Template)
	availableResources := make(map[configuration.TemplateName]map[model.DocumentType]bool)
	for _, template := range config.Templates {
		availableTemplates[template.Name] = template
		availableResources[template.Name] = make(map[model.DocumentType]bool)
		types, err := template.GetTypes()
		if err != nil {
			return nil, err
		}
		for _, t := range types {
			availableResources[template.Name][model.DocumentType(t)] = true
		}
	}
	return &Service{
		config,
		d,
		availableTemplates,
		availableResources,
		quit,
	}, nil
}

func (s *Service) ensureRequestID(w http.ResponseWriter, r *http.Request) (string, error) {
	rid := r.Header.Get(httpHeaderXRequestID)
	if rid == "" {
		rid = uuid.New()
	}
	w.Header().Del(httpHeaderXRequestID)
	w.Header().Add(httpHeaderXRequestID, rid)

	return rid, nil
}

// ListenAndServe - Blocks the current goroutine, opens an HTTP port and serves the web REST requests
func (s *Service) ListenAndServe() {

	http.HandleFunc("/espipe/health/", s.handleHealthCheck)
	http.HandleFunc("/espipe/", s.handleRequests)

	fmt.Printf("opening espipe at %v\n", s.config.EndPoint)

	s.quit <- http.ListenAndServe(s.config.EndPoint, nil)
}

// GET /espipe/health
func (s *Service) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// POST /espipe/{template}/{type}
func (s *Service) handleRequests(w http.ResponseWriter, r *http.Request) {

	rid, err := s.ensureRequestID(w, r)
	if err != nil {
		rid = ""
	}

	if r.Method != http.MethodPost {
		s.serveError(w, r, errors.ErrWrongMethod, rid)
	}

	urlSplit := strings.Split(strings.Trim(strings.ToLower(r.URL.Path), "/"), "/")
	if len(urlSplit) != 3 {
		s.serveError(w, r, errors.ErrPathNotFound, rid)
		return
	}

	templateName := configuration.TemplateName(urlSplit[1])
	template, ok := s.availableTemplates[templateName]
	if !ok {
		s.serveError(w, r, errors.ErrPathNotFound, rid)
		return
	}
	docType := model.DocumentType(urlSplit[2])
	if _, ok := s.availableResources[template.Name][docType]; !ok {
		s.serveError(w, r, errors.ErrPathNotFound, rid)
		return
	}

	// NO ERRORS -> DISPATCH
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.serveError(w, r, err, rid)
		return
	}

	document, err := model.NewDocument(&template, docType, reqBody)
	if err != nil {
		s.serveError(w, r, err, rid)
		return
	}

	s.Dispatcher.Dispatch(document)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte{})
	if err != nil {
		s.serveError(w, r, err, rid)
		return
	}
}

func (s *Service) serveError(w http.ResponseWriter, r *http.Request, err error, rid string) {

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
