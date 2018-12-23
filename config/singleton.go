package config

import (
	"io/ioutil"
	"sync"

	"gopkg.in/yaml.v2"
)

var (
	singleton *Config
	path      = "/etc/espipe/config.yaml"
	mut       = sync.RWMutex{}
)

// Set the config
func Set(configPath string) {
	mut.Lock()
	defer mut.Unlock()
	singleton = nil
	path = configPath
}

// Get the config
func Get() (config *Config, err error) {
	config = getSingleton()
	if config != nil {
		return config, nil
	}
	mut.Lock()
	defer mut.Unlock()
	config, err = loadConfig(path)
	if err != nil {
		return nil, err
	}
	singleton = config
	return config, nil
}

func loadConfig(configFile string) (*Config, error) {
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

func getSingleton() *Config {
	mut.RLock()
	defer mut.RUnlock()
	return singleton
}
