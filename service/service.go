package service

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	configuration "github.com/khezen/espipe/configuration"
	dispatcher "github.com/khezen/espipe/dispatcher"
	errors "github.com/khezen/espipe/errors"
	model "github.com/khezen/espipe/model"
)

const endpoint = ":5000"

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

// ListenAndServe - Blocks the current goroutine, opens an HTTP port and serves the web REST requests
func (s *Service) ListenAndServe() {
	http.HandleFunc("/espipe/health/", s.handleHealthCheck)
	http.HandleFunc("/espipe/", s.handleRequests)
	fmt.Printf("opening espipe at %v\n", endpoint)
	s.quit <- http.ListenAndServe(endpoint, nil)
}

// GET /espipe/health
func (s *Service) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("OK"))
}

// POST /espipe/{template}/{type}
func (s *Service) handleRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.serveError(w, r, errors.ErrWrongMethod)
	}
	urlSplit := strings.Split(strings.Trim(strings.ToLower(r.URL.Path), "/"), "/")
	if len(urlSplit) != 3 {
		s.serveError(w, r, errors.ErrPathNotFound)
		return
	}
	templateName := configuration.TemplateName(urlSplit[1])
	template, ok := s.availableTemplates[templateName]
	if !ok {
		s.serveError(w, r, errors.ErrPathNotFound)
		return
	}
	docType := model.DocumentType(urlSplit[2])
	if _, ok := s.availableResources[template.Name][docType]; !ok {
		s.serveError(w, r, errors.ErrPathNotFound)
		return
	}
	// NO ERRORS -> DISPATCH
	reqBody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.serveError(w, r, err)
		return
	}
	document, err := model.NewDocument(&template, docType, reqBody)
	if err != nil {
		s.serveError(w, r, err)
		return
	}
	s.Dispatcher.Dispatch(document)
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write([]byte{})
	if err != nil {
		s.serveError(w, r, err)
		return
	}
}
