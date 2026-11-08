package verification_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
	"github.com/lhbelfanti/ditto/v2/verification"
)

func TestNewClient_success(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusFound, Header: http.Header{"Location": {"http://example.test/next"}}}
	client := verification.NewClient(time.Second)
	client.HTTPClient.Transport = _http.MockSequenceRoundTripper(mockResult)

	want := _http.Response{Status: "302 Found", StatusCode: http.StatusFound, Header: http.Header{"Location": {"http://example.test/next"}}}
	got, _ := client.NewRequest(http.MethodGet, "http://example.test").Call(context.Background())

	assert.Equal(t, want, got)
}
