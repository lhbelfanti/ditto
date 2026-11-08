package verification_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
	"github.com/lhbelfanti/ditto/v2/verification"
)

func TestDrive_success(t *testing.T) {
	mockHeader := http.Header{"Content-Type": {"application/json"}}
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Header: mockHeader, Body: `{"id":42,"token":"abc"}`}
	mockClient := verification.MockClient(mockResult)
	mockJourneys := []verification.Journey{verification.MockJourney("account.get", "/accounts/42/v1", http.StatusOK)}

	drive := verification.MakeDrive(mockClient, "http://app/")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.get",
		Steps: []verification.Step{{
			Want: http.StatusOK,
			Got:  http.StatusOK,
			Exchange: verification.Exchange{
				Method:         http.MethodGet,
				URL:            "http://app/accounts/42/v1",
				ResponseHeader: map[string]string{"Content-Type": "application/json"},
				ResponseBody:   map[string]any{"id": float64(42), "token": "<redacted>"},
			},
		}},
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_successWhenTheRequestHasAHeaderATokenAndABody(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK}
	mockClient := verification.MockClient(mockResult)
	mockBody := map[string]string{"name": "Ana", "password": "pw"}
	mockJourneys := []verification.Journey{verification.MockSendingJourney("account.update", "/accounts/me/v1", mockBody)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.update",
		Steps: []verification.Step{{
			Want: http.StatusOK,
			Got:  http.StatusOK,
			Exchange: verification.Exchange{
				Method:        http.MethodPut,
				URL:           "http://app/accounts/me/v1",
				RequestHeader: map[string]string{"Accept": "application/json", "Authorization": "<redacted>", "Content-Type": "application/json"},
				RequestBody:   map[string]any{"name": "Ana", "password": "<redacted>"},
			},
		}},
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_successWhenAJourneySetsItselfUp(t *testing.T) {
	mockHeader := http.Header{"Location": {"http://google-mock/authorize?state=abc"}}
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusFound, Header: mockHeader}
	mockClient := verification.MockClient(mockResult)
	mockJourneys := []verification.Journey{verification.MockSetupJourney("signin", "/auth/google/v1", http.StatusFound)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "signin",
		Steps: []verification.Step{{
			Want:  http.StatusFound,
			Got:   http.StatusFound,
			Setup: true,
			Exchange: verification.Exchange{
				Method:         http.MethodGet,
				URL:            "http://app/auth/google/v1",
				ResponseHeader: map[string]string{"Location": "http://google-mock/authorize?state=%3Credacted%3E"},
			},
		}},
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_successWhenAJourneyDecodesItsResponse(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: `{"id":42}`}
	mockClient := verification.MockClient(mockResult)
	var mockTarget map[string]any
	mockJourneys := []verification.Journey{verification.MockDecodingJourney("account.get", "/accounts/42/v1", &mockTarget)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.get",
		Steps: []verification.Step{{
			Want: http.StatusOK,
			Got:  http.StatusOK,
			Exchange: verification.Exchange{
				Method:       http.MethodGet,
				URL:          "http://app/accounts/42/v1",
				ResponseBody: map[string]any{"id": float64(42)},
			},
		}},
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_successWhenTheJourneyAfterAFailedOneRuns(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK}
	mockClient := verification.MockClient(mockResult)
	mockErr := errors.New("boom")
	mockJourneys := []verification.Journey{
		verification.MockFailingJourney("session.revoke", mockErr),
		verification.MockJourney("health.ping", "http://other/ping/v1", http.StatusOK),
	}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{
		{Journey: "session.revoke", Failure: mockErr},
		{
			Journey: "health.ping",
			Steps: []verification.Step{{
				Want:     http.StatusOK,
				Got:      http.StatusOK,
				Exchange: verification.Exchange{Method: http.MethodGet, URL: "http://other/ping/v1"},
			}},
		},
	}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenAStepAnswersAnotherStatus(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusNotFound}
	mockClient := verification.MockClient(mockResult)
	mockJourneys := []verification.Journey{verification.MockJourney("account.get", "/accounts/42/v1", http.StatusOK)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.get",
		Steps: []verification.Step{{
			Want:     http.StatusOK,
			Got:      http.StatusNotFound,
			Exchange: verification.Exchange{Method: http.MethodGet, URL: "http://app/accounts/42/v1"},
		}},
		Failure: fmt.Errorf("%w: GET http://app/accounts/42/v1 answered 404, want 200", verification.ErrUnexpectedStatus),
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenTheRequestCantBeSent(t *testing.T) {
	mockErr := errors.New("boom")
	mockResult := _http.MockHTTPResult{Err: mockErr}
	mockClient := verification.MockClient(mockResult)
	mockJourneys := []verification.Journey{verification.MockJourney("account.get", "/accounts/42/v1", http.StatusOK)}
	mockURLErr := &url.Error{Op: "Get", URL: "http://app/accounts/42/v1", Err: mockErr}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.get",
		Steps:   []verification.Step{{Want: http.StatusOK, Exchange: verification.Exchange{Method: http.MethodGet, URL: "http://app/accounts/42/v1"}}},
		Failure: fmt.Errorf("%w: %w", verification.ErrCantSendRequest, fmt.Errorf("%w: %w", _http.ErrFailedToExecuteRequest, mockURLErr)),
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenTheBodyIsNotJSON(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: "pong"}
	mockClient := verification.MockClient(mockResult)
	var mockTarget map[string]any
	mockJourneys := []verification.Journey{verification.MockDecodingJourney("health.ping", "/ping/v1", &mockTarget)}
	mockDecodeErr := verification.MockDecodeError("pong")

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "health.ping",
		Steps: []verification.Step{{
			Want:     http.StatusOK,
			Got:      http.StatusOK,
			Exchange: verification.Exchange{Method: http.MethodGet, URL: "http://app/ping/v1"},
		}},
		Failure: fmt.Errorf("%w: %w", verification.ErrCantDecodeBody, mockDecodeErr),
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenARequirementIsNotMet(t *testing.T) {
	mockClient := verification.MockClient()
	mockJourneys := []verification.Journey{verification.MockRequiringJourney("session.list", false, "the second sign-in is not listed")}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "session.list",
		Failure: fmt.Errorf("%w: %s", verification.ErrRequirementNotMet, "the second sign-in is not listed"),
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenAJourneyPanics(t *testing.T) {
	mockClient := verification.MockClient()
	mockJourneys := []verification.Journey{verification.MockPanickingJourney("session.list", "index out of range")}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "session.list",
		Failure: fmt.Errorf("%w: %v", verification.ErrJourneyPanicked, "index out of range"),
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}
