package consumer

import (
	"github.com/bulklog/bulklog/consumer/elastic"
)

// Config -
type Config struct {
	Elastic *elastic.Config `yaml:"elasticsearch,omitempty"`
}

// NewConsumers -
func NewConsumers(cfg *Config) (map[string]Interface, error) {
	consumers := make(map[string]Interface)
	if cfg.Elastic != nil {
		elasticsearch := elastic.New(*cfg.Elastic)
		consumers["elasticsearch"] = elasticsearch
	}
	return consumers, nil
}
