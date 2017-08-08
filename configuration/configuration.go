package configuration

import (
	"encoding/json"
	"errors"
	"io/ioutil"
)

// Configuration contains all configuration for the logger
type Configuration struct {
	EndPoint      string        `json:"endpoint"`
	Elasticsearch string        `json:"elasticsearch"`
	Templates     []Template    `json:"templates"`
	AWSAuth       *AWSAuth      `json:"AWSAuth,omitempty"`
	BasicAuth     *Crendentials `json:"basicAuth,omitempty"`
}

// Crendentials - username & password for HTTP Basic Auth
type Crendentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AWSAuth provide credential for AWS services signing
type AWSAuth struct {
	AccessKeyID     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	Region          string `json:"region"`
}

// Template descrbies an elasticsearch Template
type Template struct {
	Name         TemplateName `json:"name"`
	BufferSizeKB float64      `json:"bufferSizeKB"`
	BufferLen    int          `json:"bufferLen"`
	TimerMS      float64      `json:"timerMS"`
	Body         interface{}  `json:"body"`
}

// TemplateName is the name of an Template
type TemplateName string

// GetTypes return declared for the given Template
func (Template *Template) GetTypes() ([]string, error) {
	body := Template.Body.(map[string]interface{})
	typesMap, ok := body["mappings"].(map[string]interface{})
	if !ok {
		return nil, errors.New("Expected map[string]interface{}. Got something else.")
	}
	types := make([]string, 0, len(typesMap))
	for t := range typesMap {
		types = append(types, t)
	}
	return types, nil
}

// LoadConfig reads the configuration from the config JSON file
func LoadConfig(configFile string) (Configuration, error) {

	// Load config file
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return Configuration{}, err
	}

	var config Configuration
	err = json.Unmarshal(bytes, &config)
	if err != nil {
		return Configuration{}, err
	}

	return config, nil
}
