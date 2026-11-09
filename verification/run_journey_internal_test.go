package verification

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

func TestRunJourney_success(t *testing.T) {
	mockSetupResult := _http.MockHTTPResult{StatusCode: http.StatusFound}
	mockStepResult := _http.MockHTTPResult{StatusCode: http.StatusOK}
	mockClient := MockClient(mockSetupResult, mockStepResult)
	recording, recordingTransport := newRecordingClient(mockClient)
	verifier := &Verifier{ctx: context.Background(), client: recording, recorder: recordingTransport, baseURL: "http://app", result: Result{Journey: "health.ping"}}
	mockStep := Step{Description: "pings", Method: http.MethodGet, Path: "/ping/v1", Status: http.StatusOK}
	mockJourney := MockSettingUpJourney("/auth/google/v1", http.StatusFound, mockStep)
	done := make(chan struct{})

	go runJourney(verifier, mockJourney, done)
	<-done

	want := &Verifier{
		ctx:      context.Background(),
		client:   recording,
		recorder: recordingTransport,
		baseURL:  "http://app",
		result: Result{
			Journey: "health.ping",
			Records: []Record{
				{Want: http.StatusFound, Got: http.StatusFound, Setup: true, Exchange: Exchange{Method: http.MethodGet, URL: "http://app/auth/google/v1"}},
				{Description: "pings", Want: http.StatusOK, Got: http.StatusOK, Exchange: Exchange{Method: http.MethodGet, URL: "http://app/ping/v1"}},
			},
		},
	}
	got := verifier

	assert.Equal(t, want, got)
}

func TestRunJourney_successWhenTheJourneyPanics(t *testing.T) {
	mockJourney := MockPanickingJourney("boom")
	verifier := &Verifier{result: Result{Journey: "health.ping"}}
	done := make(chan struct{})

	go runJourney(verifier, mockJourney, done)
	<-done

	want := &Verifier{result: Result{Journey: "health.ping", Failure: fmt.Errorf("%w: %v", ErrJourneyPanicked, "boom")}}
	got := verifier

	assert.Equal(t, want, got)
}

func TestRecordPanic_success(t *testing.T) {
	verifier := &Verifier{result: Result{Journey: "health.ping"}}

	recordPanic(verifier)

	want := &Verifier{result: Result{Journey: "health.ping"}}
	got := verifier

	assert.Equal(t, want, got)
}
