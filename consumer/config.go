package consumer

import (
	"github.com/khezen/bulklog/consumer/elastic"
)

// Config -
type Config struct {
	Elastic *elastic.Config `yaml:"elasticsearch,omitempty"`
}

// NewConsumers -
func NewConsumers(cfg *Config) ([]Interface, error) {
	consumers := make([]Interface, 0, 5)
	if cfg.Elastic != nil {
		elasticsearch := elastic.New(*cfg.Elastic)
		consumers = append(consumers, elasticsearch)
	}
	return consumers, nil
}
