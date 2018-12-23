package elastic

import "github.com/khezen/espipe/auth"

// Config -
type Config struct {
	Enabled   bool              `yaml:"enabled"`
	Address   string            `yaml:"addr"`
	Shards    int               `yaml:"shards"`
	AWSAuth   *auth.AWSConfig   `yaml:"aws_auth,omitempty"`
	BasicAuth *auth.BasicConfig `yaml:"basic_auth,omitempty"`
}
