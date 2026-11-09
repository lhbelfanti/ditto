package verification

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

func TestNewRecordingClient_success(t *testing.T) {
	mockTransport := _http.MockSequenceRoundTripper()
	client := &_http.CustomClient{HTTPClient: &http.Client{Timeout: time.Second, Transport: mockTransport}}

	recording, recordingTransport := newRecordingClient(client)

	want := []any{
		&_http.CustomClient{HTTPClient: &http.Client{Timeout: time.Second, Transport: &recorder{next: mockTransport}}},
		&recorder{next: mockTransport},
	}
	got := []any{recording, recordingTransport}

	assert.Equal(t, want, got)
}

func TestNewRecordingClient_successWhenTheClientHasNoTransport(t *testing.T) {
	client := &_http.CustomClient{HTTPClient: &http.Client{}}

	want := &recorder{next: http.DefaultTransport}
	_, got := newRecordingClient(client)

	assert.Equal(t, want, got)
}

func TestRequestBodyOf_success(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://app/sessions/v1", strings.NewReader(`{"name":"Ana"}`))

	want := []byte(`{"name":"Ana"}`)
	got := requestBodyOf(req)

	assert.Equal(t, want, got)
}

func TestRequestBodyOf_successWhenThereIsNoBody(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://app/ping/v1", nil)

	got := requestBodyOf(req)

	assert.Nil(t, got)
}

func TestResponseBodyOf_success(t *testing.T) {
	resp := &http.Response{Body: io.NopCloser(strings.NewReader("pong"))}

	want := []byte("pong")
	got := responseBodyOf(resp)

	assert.Equal(t, want, got)
}

func TestResponseBodyOf_successWhenThereIsNoBody(t *testing.T) {
	resp := &http.Response{}

	got := responseBodyOf(resp)

	assert.Nil(t, got)
}
