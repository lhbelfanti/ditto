package verification

import (
	"encoding/json"
	"net/http"
	"time"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

type (
	// MockedJourney is a journey whose Setup runs setup and whose Steps are steps.
	MockedJourney struct {
		setup MockSetup
		steps []Step
	}

	// MockSetup is what the Setup of a MockedJourney does.
	MockSetup func(verifier *Verifier)
)

// Setup runs the setup of the mock, when it has one.
func (m *MockedJourney) Setup(verifier *Verifier) {
	if m.setup != nil {
		m.setup(verifier)
	}
}

// Steps returns the steps of the mock.
func (m *MockedJourney) Steps() []Step {
	return m.steps
}

// MockClient returns the client NewClient creates, with its requests answered by results in sequence.
func MockClient(results ..._http.MockHTTPResult) *_http.CustomClient {
	client := NewClient(time.Second)
	client.HTTPClient.Transport = _http.MockSequenceRoundTripper(results...)

	return client
}

// MockJourney returns a journey without setup that sends steps.
func MockJourney(steps ...Step) *MockedJourney {
	return &MockedJourney{steps: steps}
}

// MockSettingUpJourney returns a journey whose Setup sends GET path expecting status, and that then sends steps.
func MockSettingUpJourney(path string, status int, steps ...Step) *MockedJourney {
	return &MockedJourney{
		setup: func(verifier *Verifier) {
			verifier.Send(verifier.NewRequest(http.MethodGet, verifier.URL(path)), status)
		},
		steps: steps,
	}
}

// MockDecodingJourney returns a journey whose Setup sends GET path, expecting a 200, and decodes its body into
// target.
func MockDecodingJourney(path string, target any) *MockedJourney {
	return &MockedJourney{setup: func(verifier *Verifier) {
		verifier.Decode(verifier.Send(verifier.NewRequest(http.MethodGet, verifier.URL(path)), http.StatusOK), target)
	}}
}

// MockPanickingJourney returns a journey whose Setup panics with value.
func MockPanickingJourney(value any) *MockedJourney {
	return &MockedJourney{setup: func(_ *Verifier) {
		panic(value)
	}}
}

// MockFailingJourney returns a journey whose Setup fails with err.
func MockFailingJourney(err error) *MockedJourney {
	return &MockedJourney{setup: func(verifier *Verifier) {
		verifier.fail(err)
	}}
}

// MockDecodeError returns the error encoding/json gives for body when it is not JSON.
func MockDecodeError(body string) error {
	var target any
	return json.Unmarshal([]byte(body), &target)
}

// MockValues returns path values that are always values.
func MockValues(values map[string]string) Values {
	return func() map[string]string {
		return values
	}
}

// MockWithHeader returns a With that sets the header key to value.
func MockWithHeader(key, value string) With {
	return func(request *_http.Request) {
		request.WithHeader(key, value)
	}
}

// MockChecker returns a Checker that answers err for any data.
func MockChecker[T any](err error) Checker[T] {
	return func(_ T) error {
		return err
	}
}

// MockCard returns a feature card with the journey session.revoke-one, whose table is table, followed by a
// journey that is not driven, with a table of its own.
func MockCard(table string) string {
	return "# Sessions\n\n" +
		"### session.revoke-one — Revoke one session\n\n" +
		"Setup: two sessions of one person.\n\n" +
		table + "\n" +
		"### session.other — Not driven\n\n" +
		"| # | Request | Status | Proves |\n|---|---|---|---|\n| 1 | `GET /other/v1` | 200 | stays |\n"
}
