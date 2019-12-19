package auth

import (
	"encoding/base64"
	"fmt"
	"net/http"
)

// BasicConfig - username & password for HTTP Basic Auth
type BasicConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// BasicAuthSigner provides basic http authentication
type basicSigner struct {
	username, password string
	authorization      string
}

// NewBasicSigner provides basic authentication to given request
func NewBasicSigner(cfg BasicConfig) Signer {
	creds := fmt.Sprintf("%s:%s", cfg.Username, cfg.Password)
	token := base64.RawURLEncoding.EncodeToString([]byte(creds))
	authorization := fmt.Sprintf("Basic %s", token)
	return &basicSigner{
		cfg.Username, cfg.Password, authorization,
	}
}

// Sign sign the request with basic autnetication
func (s *basicSigner) Sign(r *http.Request, body []byte) error {
	r.Header.Set("Authorization", s.authorization)
	return nil
}
