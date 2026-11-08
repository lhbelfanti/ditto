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

// MockRequest returns a request for GET /accounts/{id}/v1 with the id 42, that expects a 200.
func MockRequest() Request {
	return MockRequestWith("GET /accounts/{id}/v1", http.StatusOK)
}

// MockRequestWith returns a request for route with the id 42, that expects status.
func MockRequestWith(route string, status int) Request {
	return Request{
		Route:  route,
		Params: map[string]string{"id": "42"},
		Status: status,
	}
}

// MockJourney returns a journey that sends each request as a step.
func MockJourney(id string, requests ...Request) Journey {
	return Journey{ID: id, Prove: func(run *Run) {
		for _, request := range requests {
			run.Do(request)
		}
	}}
}

// MockSetupJourney returns a journey that sends request to prepare itself, and then nothing else.
func MockSetupJourney(id string, request Request) Journey {
	return Journey{ID: id, Prove: func(run *Run) {
		run.Send(request)
	}}
}

// MockDecodingJourney returns a journey that sends request as a step and decodes its body into target.
func MockDecodingJourney(id string, request Request, target any) Journey {
	return Journey{ID: id, Prove: func(run *Run) {
		response := run.Do(request)
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
