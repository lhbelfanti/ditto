package verification

import (
	"net/http"
	"time"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

// NewClient creates a client whose requests time out after timeout, and that answers a redirect with the
// redirect itself instead of following it, so that a journey can check where it points.
func NewClient(timeout time.Duration) *_http.CustomClient {
	return &_http.CustomClient{HTTPClient: &http.Client{Timeout: timeout, CheckRedirect: stopAtRedirect}}
}

// stopAtRedirect makes the client return the redirect response instead of following it.
func stopAtRedirect(_ *http.Request, _ []*http.Request) error {
	return http.ErrUseLastResponse
}
