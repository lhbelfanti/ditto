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
