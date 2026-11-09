package verification

import (
	"encoding/json"
	"net/http"
	"time"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

// MockClient returns the client NewClient creates, with its requests answered by results in sequence.
func MockClient(results ..._http.MockHTTPResult) *_http.CustomClient {
	client := NewClient(time.Second)
	client.HTTPClient.Transport = _http.MockSequenceRoundTripper(results...)

	return client
}

// MockJourney returns a journey that sends GET path as a step and expects status.
func MockJourney(id string, path string, status int) Journey {
	return Journey{ID: id, Prove: func(run *Run) {
		run.Do(run.NewRequest(http.MethodGet, run.URL(path)), status)
	}}
}

// MockSetupJourney returns a journey that sends GET path to prepare itself, expecting status, and then nothing else.
func MockSetupJourney(id string, path string, status int) Journey {
	return Journey{ID: id, Prove: func(run *Run) {
		run.Setup(run.NewRequest(http.MethodGet, run.URL(path)), status)
	}}
}

// MockSendingJourney returns a journey that sends PUT path as a step with a bearer token, an Accept header and
// body, and expects a 200.
func MockSendingJourney(id string, path string, body any) Journey {
	return Journey{ID: id, Prove: func(run *Run) {
		request := run.NewRequest(http.MethodPut, run.URL(path))
		request.WithHeader("Authorization", "Bearer session")
		request.WithHeader("Accept", "application/json")
		request.WithBody(body)
		run.Do(request, http.StatusOK)
	}}
}

// MockDecodingJourney returns a journey that sends GET path as a step, expecting a 200, and decodes its body into
// target.
func MockDecodingJourney(id string, path string, target any) Journey {
	return Journey{ID: id, Prove: func(run *Run) {
		response := run.Do(run.NewRequest(http.MethodGet, run.URL(path)), http.StatusOK)
		run.Decode(response, target)
	}}
}

// MockRequiringJourney returns a journey that requires condition, with reason.
func MockRequiringJourney(id string, condition bool, reason string) Journey {
	return Journey{ID: id, Prove: func(run *Run) {
		run.Require(condition, reason)
	}}
}

// MockPanickingJourney returns a journey that panics with value.
func MockPanickingJourney(id string, value any) Journey {
	return Journey{ID: id, Prove: func(_ *Run) {
		panic(value)
	}}
}

// MockFailingJourney returns a journey that fails with err.
func MockFailingJourney(id string, err error) Journey {
	return Journey{ID: id, Prove: func(run *Run) {
		run.fail(err)
	}}
}

// MockDecodeError returns the error encoding/json gives for body when it is not JSON.
func MockDecodeError(body string) error {
	var target any
	return json.Unmarshal([]byte(body), &target)
}
