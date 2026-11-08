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
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: `{"id":42,"token":"secret"}`}
	mockClient := verification.MockClient(mockResult)
	mockRequest := verification.MockRequest()
	mockJourneys := []verification.Journey{verification.MockJourney("account.get", mockRequest)}

	drive := verification.MakeDrive(mockClient, "http://app/")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.get",
		Steps: []verification.Step{{
			Route: "GET /accounts/{id}/v1",
			Want:  http.StatusOK,
			Got:   http.StatusOK,
			Exchange: verification.Exchange{
				Method:       http.MethodGet,
				URL:          "http://app/accounts/42/v1",
				ResponseBody: map[string]any{"id": float64(42), "token": "<redacted>"},
			},
		}},
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_successWhenASetupRequestHasNoStatus(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusFound}
	mockClient := verification.MockClient(mockResult)
	mockRequest := verification.MockRequestWith("GET /accounts/{id}/v1", 0)
	mockJourneys := []verification.Journey{verification.MockSetupJourney("signin", mockRequest)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "signin",
		Steps: []verification.Step{{
			Route:    "GET /accounts/{id}/v1",
			Got:      http.StatusFound,
			Setup:    true,
			Exchange: verification.Exchange{Method: http.MethodGet, URL: "http://app/accounts/42/v1"},
		}},
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_successWhenAJourneyDecodesItsResponse(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: `{"id":42}`}
	mockClient := verification.MockClient(mockResult)
	mockRequest := verification.MockRequest()
	mockTarget := map[string]any{}
	mockJourneys := []verification.Journey{verification.MockDecodingJourney("account.get", mockRequest, &mockTarget)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.get",
		Steps: []verification.Step{{
			Route: "GET /accounts/{id}/v1",
			Want:  http.StatusOK,
			Got:   http.StatusOK,
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

func TestDrive_failsWhenAStepAnswersAnotherStatus(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusInternalServerError}
	mockClient := verification.MockClient(mockResult)
	mockFirst := verification.MockRequest()
	mockSecond := verification.MockRequestWith("DELETE /accounts/{id}/v1", http.StatusOK)
	mockJourneys := []verification.Journey{verification.MockJourney("account.get", mockFirst, mockSecond)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.get",
		Steps: []verification.Step{{
			Route:    "GET /accounts/{id}/v1",
			Want:     http.StatusOK,
			Got:      http.StatusInternalServerError,
			Exchange: verification.Exchange{Method: http.MethodGet, URL: "http://app/accounts/42/v1"},
		}},
		Failure: fmt.Errorf("%w: GET /accounts/{id}/v1 answered 500, want 200", verification.ErrUnexpectedStatus),
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenTheRequestCantBeSent(t *testing.T) {
	mockErr := errors.New("boom")
	mockResult := _http.MockHTTPResult{Err: mockErr}
	mockClient := verification.MockClient(mockResult)
	mockRequest := verification.MockRequest()
	mockJourneys := []verification.Journey{verification.MockJourney("account.get", mockRequest)}
	mockURLErr := &url.Error{Op: "Get", URL: "http://app/accounts/42/v1", Err: mockErr}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.get",
		Steps: []verification.Step{{
			Route:    "GET /accounts/{id}/v1",
			Want:     http.StatusOK,
			Exchange: verification.Exchange{Method: http.MethodGet, URL: "http://app/accounts/42/v1"},
		}},
		Failure: fmt.Errorf("%w: %w", verification.ErrCantSendRequest, fmt.Errorf("%w: %w", _http.ErrFailedToExecuteRequest, mockURLErr)),
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenTheRequestIsInvalid(t *testing.T) {
	mockClient := verification.MockClient()
	mockRequest := verification.MockRequestWith("/accounts/{id}/v1", http.StatusOK)
	mockJourneys := []verification.Journey{verification.MockJourney("account.get", mockRequest)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{Journey: "account.get", Failure: fmt.Errorf("%w: %w", verification.ErrInvalidRequest, verification.ErrRouteWithoutMethod)}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenTheBodyIsNotJSON(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: "pong"}
	mockClient := verification.MockClient(mockResult)
	mockRequest := verification.MockRequest()
	mockTarget := map[string]any{}
	mockJourneys := []verification.Journey{verification.MockDecodingJourney("health.ping", mockRequest, &mockTarget)}
	mockDecodeErr := verification.MockDecodeError("pong")

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "health.ping",
		Steps: []verification.Step{{
			Route:    "GET /accounts/{id}/v1",
			Want:     http.StatusOK,
			Got:      http.StatusOK,
			Exchange: verification.Exchange{Method: http.MethodGet, URL: "http://app/accounts/42/v1"},
		}},
		Failure: fmt.Errorf("%w: %w", verification.ErrCantDecodeBody, mockDecodeErr),
	}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenARequirementIsNotMet(t *testing.T) {
	mockClient := verification.MockClient()
	mockJourneys := []verification.Journey{verification.MockRequiringJourney("session.revoke", false, "the other session is still active")}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{Journey: "session.revoke", Failure: fmt.Errorf("%w: %s", verification.ErrRequirementNotMet, "the other session is still active")}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenAJourneyPanics(t *testing.T) {
	mockClient := verification.MockClient()
	mockJourneys := []verification.Journey{
		verification.MockPanickingJourney("session.revoke", "index out of range"),
		verification.MockRequiringJourney("profile.get", true, "unused"),
	}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{
		{Journey: "session.revoke", Failure: fmt.Errorf("%w: %v", verification.ErrJourneyPanicked, "index out of range")},
		{Journey: "profile.get"},
	}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}
