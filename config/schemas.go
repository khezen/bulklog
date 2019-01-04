package config

import (
	"github.com/khezen/bulklog/collection"
	"github.com/khezen/bulklog/consumer"
	"github.com/khezen/bulklog/redisc"
)

// Config contains all configuration for the logger
type Config struct {
	Port        int                 `yaml:"port"`
	Persistence Persistence         `yaml:"persistence"`
	Output      consumer.Config     `yaml:"output"`
	Collections []collection.Config `yaml:"collections,flow"`
}

// Persistence -
type Persistence struct {
	Enabled bool          `yaml:"enabled"`
	Redis   redisc.Config `yaml:"redis"`
}
