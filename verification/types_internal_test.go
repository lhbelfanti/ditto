package verification

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

func TestRun_send_success(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: "pong"}
	mockClient := MockClient(mockResult)
	recording, recordingTransport := newRecordingClient(mockClient)
	mockRequest := MockRequest()
	run := &Run{ctx: context.Background(), client: recording, recorder: recordingTransport, baseURL: "http://app"}

	response := run.send(mockRequest, false)

	want := []any{
		_http.Response{Body: "pong", Status: "200 OK", StatusCode: http.StatusOK},
		&Run{
			ctx:      context.Background(),
			client:   recording,
			recorder: recordingTransport,
			baseURL:  "http://app",
			result: Result{Steps: []Step{{
				Route:    "GET /accounts/{id}/v1",
				Want:     http.StatusOK,
				Got:      http.StatusOK,
				Exchange: Exchange{Method: http.MethodGet, URL: "http://app/accounts/42/v1"},
			}}},
		},
	}
	got := []any{response, run}

	assert.Equal(t, want, got)
}

func TestRun_send_successWhenTheRequestHasAHeaderATokenAndABody(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK}
	mockClient := MockClient(mockResult)
	recording, recordingTransport := newRecordingClient(mockClient)
	mockRequest := Request{
		Route:  "PUT /accounts/me/v1",
		Token:  "session",
		Header: map[string]string{"Accept": "application/json"},
		Body:   map[string]string{"name": "Ana"},
		Status: http.StatusOK,
	}
	run := &Run{ctx: context.Background(), client: recording, recorder: recordingTransport, baseURL: "http://app"}

	run.send(mockRequest, false)

	want := &Run{
		ctx:      context.Background(),
		client:   recording,
		recorder: recordingTransport,
		baseURL:  "http://app",
		result: Result{Steps: []Step{{
			Route: "PUT /accounts/me/v1",
			Want:  http.StatusOK,
			Got:   http.StatusOK,
			Exchange: Exchange{
				Method:        http.MethodPut,
				URL:           "http://app/accounts/me/v1",
				RequestHeader: map[string]string{"Accept": "application/json", "Authorization": redacted, "Content-Type": "application/json"},
				RequestBody:   map[string]any{"name": "Ana"},
			},
		}}},
	}
	got := run

	assert.Equal(t, want, got)
}

func TestRun_send_failsWhenASetupStepAnswersAnotherStatus(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusFound}
	mockClient := MockClient(mockResult)
	recording, recordingTransport := newRecordingClient(mockClient)
	mockRequest := MockRequest()
	mockJourney := MockSetupJourney("signin", mockRequest)
	run := &Run{ctx: context.Background(), client: recording, recorder: recordingTransport, baseURL: "http://app"}
	done := make(chan struct{})

	go runJourney(run, mockJourney.Prove, done)
	<-done

	want := &Run{
		ctx:      context.Background(),
		client:   recording,
		recorder: recordingTransport,
		baseURL:  "http://app",
		result: Result{
			Steps: []Step{{
				Route:    "GET /accounts/{id}/v1",
				Want:     http.StatusOK,
				Got:      http.StatusFound,
				Setup:    true,
				Exchange: Exchange{Method: http.MethodGet, URL: "http://app/accounts/42/v1"},
			}},
			Failure: fmt.Errorf("%w: GET /accounts/{id}/v1 answered 302, want 200", ErrUnexpectedStatus),
		},
	}
	got := run

	assert.Equal(t, want, got)
}

func TestRun_fail_success(t *testing.T) {
	mockErr := errors.New("boom")
	mockJourney := MockFailingJourney("account.get", mockErr)
	run := &Run{}
	done := make(chan struct{})

	go runJourney(run, mockJourney.Prove, done)
	<-done

	want := &Run{result: Result{Failure: mockErr}}
	got := run

	assert.Equal(t, want, got)
}

func TestRecorder_RoundTrip_success(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: `{"id":1}`}
	mockTransport := _http.MockSequenceRoundTripper(mockResult)
	transport := &recorder{next: mockTransport}
	req, _ := http.NewRequest(http.MethodPost, "http://app/sessions/v1?code=xyz", strings.NewReader(`{"name":"Ana"}`))

	resp, _ := transport.RoundTrip(req)

	want := []any{
		`{"id":1}`,
		Exchange{
			Method:       http.MethodPost,
			URL:          "http://app/sessions/v1?code=%3Credacted%3E",
			RequestBody:  map[string]any{"name": "Ana"},
			ResponseBody: map[string]any{"id": float64(1)},
		},
	}
	got := []any{string(responseBodyOf(resp)), transport.take()}

	assert.Equal(t, want, got)
}

func TestRecorder_RoundTrip_failsWhenTheTransportFails(t *testing.T) {
	mockErr := errors.New("boom")
	mockResult := _http.MockHTTPResult{Err: mockErr}
	mockTransport := _http.MockSequenceRoundTripper(mockResult)
	transport := &recorder{next: mockTransport}
	req, _ := http.NewRequest(http.MethodGet, "http://app/ping/v1", nil)

	want := mockErr
	_, got := transport.RoundTrip(req)

	assert.ErrorIs(t, got, want)
}

func TestRecorder_keep_success(t *testing.T) {
	transport := &recorder{}

	transport.keep(Exchange{Method: http.MethodGet, URL: "http://app/ping/v1"})

	want := &recorder{exchange: Exchange{Method: http.MethodGet, URL: "http://app/ping/v1"}}
	got := transport

	assert.Equal(t, want, got)
}

func TestRecorder_take_success(t *testing.T) {
	transport := &recorder{exchange: Exchange{Method: http.MethodGet, URL: "http://app/ping/v1"}}

	want := []any{Exchange{Method: http.MethodGet, URL: "http://app/ping/v1"}, &recorder{}}
	got := []any{transport.take(), transport}

	assert.Equal(t, want, got)
}
