package elastic

import "github.com/khezen/bulklog/pkg/auth"

// Config -
type Config struct {
	Enabled   bool              `yaml:"enabled"`
	Endpoint  string            `yaml:"endpoint"`
	Scheme    string            `yaml:"scheme"`
	AWSAuth   *auth.AWSConfig   `yaml:"aws_auth,omitempty"`
	BasicAuth *auth.BasicConfig `yaml:"basic_auth,omitempty"`
}
