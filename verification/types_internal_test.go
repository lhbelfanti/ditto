package verification

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

func TestRun_send_success(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: "pong"}
	mockClient, _ := _http.MockSequenceClient(mockResult)
	mockRequest := MockRequest()
	run := &Run{ctx: context.Background(), client: mockClient, baseURL: "http://app"}

	response := run.send(mockRequest, false)

	want := []any{http.StatusOK, "pong", []Step{{Route: "GET /accounts/{id}/v1", Want: http.StatusOK, Got: http.StatusOK, Exchange: Exchange{URL: "http://app/accounts/42/v1"}}}}
	got := []any{response.StatusCode, response.Body, run.result.Steps}

	assert.Equal(t, want, got)
}

func TestRun_send_successWhenTheRequestHasAHeaderATokenAndABody(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK}
	mockClient, _ := _http.MockSequenceClient(mockResult)
	mockRequest := Request{Route: "PUT /accounts/me/v1", Token: "session", Header: map[string]string{"Accept": "application/json"}, Body: map[string]string{"name": "Ana"}, Status: http.StatusOK}
	run := &Run{ctx: context.Background(), client: mockClient, baseURL: "http://app"}

	run.send(mockRequest, false)

	want := Exchange{
		URL:           "http://app/accounts/me/v1",
		RequestHeader: map[string]string{"Accept": "application/json", "Authorization": redacted},
		RequestBody:   map[string]any{"name": "Ana"},
	}
	got := run.result.Steps[0].Exchange

	assert.Equal(t, want, got)
}

func TestRun_send_failsWhenASetupStepAnswersAnotherStatus(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusFound}
	mockClient, _ := _http.MockSequenceClient(mockResult)
	mockRequest := MockRequest()
	run := &Run{ctx: context.Background(), client: mockClient, baseURL: "http://app"}
	done := make(chan struct{})

	go func() { defer close(done); run.send(mockRequest, true) }()
	<-done

	want := "unexpected status: GET /accounts/{id}/v1 answered 302, want 200"
	got := run.result.Failure

	assert.Equal(t, want, got)
}

func TestRun_fail_success(t *testing.T) {
	mockErr := errors.New("boom")
	run := &Run{}
	done := make(chan struct{})

	go func() { defer close(done); run.fail(mockErr) }()
	<-done

	want := "boom"
	got := run.result.Failure

	assert.Equal(t, want, got)
}
