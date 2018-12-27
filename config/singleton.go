package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"gopkg.in/yaml.v2"
)

var (
	singleton *Config
)

// Get the config
func Get() (config *Config, err error) {
	if singleton != nil {
		return singleton, nil
	}
	singleton, err = loadConfig()
	if err != nil {
		return nil, err
	}
	return singleton, nil
}

func loadConfig() (*Config, error) {
	// where is config?
	configPath := strings.TrimRight(os.Getenv("CONFIG_PATH"), "/")
	configFile := fmt.Sprintf("%s/config.yaml", configPath)
	// Load config file
	bytes, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(bytes, &config)
	if err != nil {
		return nil, err
	}
	return &config, nil
}
