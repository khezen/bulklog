package config

import (
	"github.com/khezen/bulklog/collection"
	"github.com/khezen/bulklog/consumer/elastic"
)

// Config contains all configuration for the logger
type Config struct {
	Port        int                 `yaml:"port"`
	Redis       Redis               `yaml:"redis"`
	Output      Consumers           `yaml:"output"`
	Collections []collection.Config `yaml:"collections,flow"`
}

// Consumers -
type Consumers struct {
	Elastic *elastic.Config `yaml:"elasticsearch,omitempty"`
}

// Redis -
type Redis struct {
	Enabled  bool   `yaml:"enabled"`
	Endpoint string `yaml:"endpoint"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}
