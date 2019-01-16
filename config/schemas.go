package config

import (
	"github.com/bulklog/bulklog/collection"
	"github.com/bulklog/bulklog/output"
)

// Config contains all configuration for the logger
type Config struct {
	Port        int                 `yaml:"port"`
	Persistence Persistence         `yaml:"persistence"`
	Output      output.Config       `yaml:"output"`
	Collections []collection.Config `yaml:"collections,flow"`
}

// Persistence -
type Persistence struct {
	Enabled bool  `yaml:"enabled"`
	Redis   Redis `yaml:"redis"`
}

// Redis - redis config
type Redis struct {
	Endpoint string `yaml:"endpoint"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
	IdleConn int    `yaml:"idle_conn"`
	MaxConn  int    `yaml:"max_conn"`
}
