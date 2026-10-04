package http

import (
	"net/http"
	"time"
)

// NewClient create a new CustomClient
func NewClient(timeout time.Duration) *CustomClient {
	return &CustomClient{
		HTTPClient: &http.Client{Timeout: timeout},
	}
}
