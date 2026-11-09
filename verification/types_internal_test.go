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

func TestStep_path_success(t *testing.T) {
	mockPathValues := map[string]string{"account": "7", "id": "42"}
	step := Step{Path: "/accounts/{account}/sessions/{id}/v1", PathValues: MockValues(mockPathValues)}

	want := "/accounts/7/sessions/42/v1"
	got := step.path()

	assert.Equal(t, want, got)
}

func TestStep_path_successWhenTheStepHasNoPathValues(t *testing.T) {
	step := Step{Path: "/ping/v1"}

	want := "/ping/v1"
	got := step.path()

	assert.Equal(t, want, got)
}

func TestVerifier_URL_success(t *testing.T) {
	verifier := &Verifier{baseURL: "http://app/"}

	want := "http://app/accounts/42/v1"
	got := verifier.URL("/accounts/42/v1")

	assert.Equal(t, want, got)
}

func TestVerifier_URL_successWhenThePathIsAFullURL(t *testing.T) {
	verifier := &Verifier{baseURL: "http://app"}

	want := "http://google-mock/authorize"
	got := verifier.URL("http://google-mock/authorize")

	assert.Equal(t, want, got)
}

func TestVerifier_Send_success(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: "pong"}
	mockClient := MockClient(mockResult)
	recording, recordingTransport := newRecordingClient(mockClient)
	verifier := &Verifier{ctx: context.Background(), client: recording, recorder: recordingTransport, baseURL: "http://app"}

	response := verifier.Send(verifier.NewRequest(http.MethodGet, "http://app/ping/v1"), http.StatusOK)

	want := []any{
		_http.Response{Body: "pong", Status: "200 OK", StatusCode: http.StatusOK},
		&Verifier{
			ctx:      context.Background(),
			client:   recording,
			recorder: recordingTransport,
			baseURL:  "http://app",
			result:   Result{Records: []Record{{Want: http.StatusOK, Got: http.StatusOK, Setup: true, Exchange: Exchange{Method: http.MethodGet, URL: "http://app/ping/v1"}}}},
		},
	}
	got := []any{response, verifier}

	assert.Equal(t, want, got)
}

func TestVerifier_prove_success(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: `{"code":200,"data":"pong"}`}
	mockClient := MockClient(mockResult)
	recording, recordingTransport := newRecordingClient(mockClient)
	verifier := &Verifier{ctx: context.Background(), client: recording, recorder: recordingTransport, baseURL: "http://app"}
	mockStep := Step{Description: "pings", Method: http.MethodGet, Path: "/ping/v1", Status: http.StatusOK, Expect: Data("pong")}

	verifier.prove(mockStep)

	want := &Verifier{
		ctx:      context.Background(),
		client:   recording,
		recorder: recordingTransport,
		baseURL:  "http://app",
		result: Result{Records: []Record{{
			Description: "pings",
			Want:        http.StatusOK,
			Got:         http.StatusOK,
			Exchange:    Exchange{Method: http.MethodGet, URL: "http://app/ping/v1", ResponseBody: map[string]any{"code": float64(200), "data": "pong"}},
		}}},
	}
	got := verifier

	assert.Equal(t, want, got)
}

func TestVerifier_prove_failsWhenTheExpectFails(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: "pong"}
	mockClient := MockClient(mockResult)
	recording, recordingTransport := newRecordingClient(mockClient)
	verifier := &Verifier{ctx: context.Background(), client: recording, recorder: recordingTransport, baseURL: "http://app"}
	mockStep := Step{Description: "pings", Method: http.MethodGet, Path: "/ping/v1", Status: http.StatusOK, Expect: Data("pong")}
	mockJourney := MockJourney(mockStep)
	mockDecodeErr := MockDecodeError("pong")
	done := make(chan struct{})

	go runJourney(verifier, mockJourney, done)
	<-done

	want := &Verifier{
		ctx:      context.Background(),
		client:   recording,
		recorder: recordingTransport,
		baseURL:  "http://app",
		result: Result{
			Records: []Record{{Description: "pings", Want: http.StatusOK, Got: http.StatusOK, Exchange: Exchange{Method: http.MethodGet, URL: "http://app/ping/v1"}}},
			Failure: fmt.Errorf("%w: %s: %w", ErrUnexpectedResponse, "pings", fmt.Errorf("%w: %w", ErrCantDecodeBody, mockDecodeErr)),
		},
	}
	got := verifier

	assert.Equal(t, want, got)
}

func TestVerifier_send_failsWhenTheRequestAnswersAnotherStatus(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusFound}
	mockClient := MockClient(mockResult)
	mockJourney := MockSettingUpJourney("/auth/google/v1", http.StatusOK)
	recording, recordingTransport := newRecordingClient(mockClient)
	verifier := &Verifier{ctx: context.Background(), client: recording, recorder: recordingTransport, baseURL: "http://app"}
	done := make(chan struct{})

	go runJourney(verifier, mockJourney, done)
	<-done

	want := &Verifier{
		ctx:      context.Background(),
		client:   recording,
		recorder: recordingTransport,
		baseURL:  "http://app",
		result: Result{
			Records: []Record{{
				Want:     http.StatusOK,
				Got:      http.StatusFound,
				Setup:    true,
				Exchange: Exchange{Method: http.MethodGet, URL: "http://app/auth/google/v1"},
			}},
			Failure: fmt.Errorf("%w: GET http://app/auth/google/v1 answered 302, want 200", ErrUnexpectedStatus),
		},
	}
	got := verifier

	assert.Equal(t, want, got)
}

func TestVerifier_fail_success(t *testing.T) {
	mockErr := errors.New("boom")
	mockJourney := MockFailingJourney(mockErr)
	verifier := &Verifier{}
	done := make(chan struct{})

	go runJourney(verifier, mockJourney, done)
	<-done

	want := &Verifier{result: Result{Failure: mockErr}}
	got := verifier

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
