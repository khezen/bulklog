package elastic

import (
	"bytes"
	"encoding/base64"
	"net/http"
)

// BasicAuthSigner provides basic http authentication
type BasicAuthSigner struct {
	username, password string
	auth               string
}

// Sign sign the request with basic autnetication
func (s *BasicAuthSigner) Sign(r *http.Request) (http.Header, error) {
	r.Header.Add("Authorization", s.auth)
	return r.Header, nil
}

// NewBasicAuthSigner provides basic authentication to given request
func NewBasicAuthSigner(username, password string) *BasicAuthSigner {
	authBuf := bytes.NewBufferString("Basic ")
	authBuf.WriteString(username)
	authBuf.WriteString(":")
	authBuf.WriteString(password)
	token := base64.RawURLEncoding.EncodeToString(authBuf.Bytes())
	return &BasicAuthSigner{
		username, password, token,
	}
}
