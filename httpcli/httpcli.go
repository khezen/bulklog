package httpcli

import (
	"net/http"
	"time"
)

var (
	transport = &http.Transport{
		MaxIdleConns:       10,
		IdleConnTimeout:    30 * time.Second,
		DisableCompression: true,
	}
	singleton = &http.Client{
		Transport: transport,
	}
)

// Singleton return the shared httpClient
func Singleton() *http.Client {
	return singleton
}
