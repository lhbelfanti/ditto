package http

import (
	"net/http"
	"time"
)

// NewClient creates a CustomClient whose requests time out after timeout.
func NewClient(timeout time.Duration) *CustomClient {
	return &CustomClient{
		HTTPClient: &http.Client{Timeout: timeout},
	}
}

// WithHeader sets a header on the request, replacing any value it already has.
func WithHeader(key, value string) RequestOption {
	return func(req *http.Request) {
		req.Header.Set(key, value)
	}
}
