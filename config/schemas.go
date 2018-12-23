package config

import (
	"github.com/khezen/espipe/collection"
	"github.com/khezen/espipe/consumer/elastic"
)

// Config contains all configuration for the logger
type Config struct {
	Redis       Redis               `yaml:"redis"`
	Consumers   Consumers           `yaml:"consumers"`
	Collections []collection.Config `yaml:"collections,flow"`
}

// Consumers -
type Consumers struct {
	Elastic *elastic.Config `yaml:"elasticsearch,omitempty"`
}

// Redis -
type Redis struct {
	Enabled   bool   `yaml:"enabled"`
	Address   string `yaml:"address"`
	Password  string `yaml:"password"`
	Partition int    `yaml:"partition"`
}
