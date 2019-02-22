package output

import (
	"github.com/khezen/bulklog/output/elastic"
)

// Config -
type Config struct {
	Elastic *elastic.Config `yaml:"elasticsearch,omitempty"`
}

// NewOutputs -
func NewOutputs(cfg *Config) (map[string]Interface, error) {
	outputs := make(map[string]Interface)
	if cfg.Elastic != nil {
		elasticsearch := elastic.New(*cfg.Elastic)
		outputs["elasticsearch"] = elasticsearch
	}
	return outputs, nil
}
