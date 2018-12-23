package auth

import "net/http"

// Signer add the Authorization header to http request
type Signer interface {
	Sign(r *http.Request, body []byte) error
}
