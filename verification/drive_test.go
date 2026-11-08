package verification_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
	"github.com/lhbelfanti/ditto/v2/verification"
)

func TestDrive_success(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: `{"id":42,"token":"secret"}`}
	mockClient, _ := _http.MockSequenceClient(mockResult)
	mockRequest := verification.MockRequest()
	mockJourneys := []verification.Journey{verification.MockJourney("account.get", mockRequest)}

	drive := verification.MakeDrive(mockClient, "http://app/")

	want := verification.Report{
		Results: []verification.Result{
			{
				Journey: "account.get",
				Steps: []verification.Step{
					{
						Route: "GET /accounts/{id}/v1",
						Want:  http.StatusOK,
						Got:   http.StatusOK,
						Exchange: verification.Exchange{
							URL: "http://app/accounts/42/v1",
							ResponseBody: map[string]any{
								"id":    float64(42),
								"token": "<redacted>",
							},
						},
					},
				},
			},
		},
	}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_successWhenASetupRequestHasNoStatus(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusFound}
	mockClient, _ := _http.MockSequenceClient(mockResult)
	mockRequest := verification.MockRequestWith("GET /accounts/{id}/v1", 0)
	mockJourneys := []verification.Journey{verification.MockSetupJourney("signin", mockRequest)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{
		Results: []verification.Result{
			{
				Journey: "signin",
				Steps: []verification.Step{
					{
						Route: "GET /accounts/{id}/v1",
						Got:   http.StatusFound,
						Setup: true,
						Exchange: verification.Exchange{
							URL: "http://app/accounts/42/v1",
						},
					},
				},
			},
		},
	}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_successWhenAJourneyDecodesItsResponse(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: `{"id":42}`}
	mockClient, _ := _http.MockSequenceClient(mockResult)
	mockRequest := verification.MockRequest()
	mockTarget := map[string]any{}
	mockJourneys := []verification.Journey{verification.MockDecodingJourney("account.get", mockRequest, &mockTarget)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{
		Results: []verification.Result{
			{
				Journey: "account.get",
				Steps: []verification.Step{
					{
						Route: "GET /accounts/{id}/v1",
						Want:  http.StatusOK,
						Got:   http.StatusOK,
						Exchange: verification.Exchange{
							URL: "http://app/accounts/42/v1",
							ResponseBody: map[string]any{
								"id": float64(42),
							},
						},
					},
				},
			},
		},
	}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenAStepAnswersAnotherStatus(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusInternalServerError}
	mockClient, _ := _http.MockSequenceClient(mockResult)
	mockFirst := verification.MockRequest()
	mockSecond := verification.MockRequestWith("DELETE /accounts/{id}/v1", http.StatusOK)
	mockJourneys := []verification.Journey{verification.MockJourney("account.get", mockFirst, mockSecond)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{
		Results: []verification.Result{
			{
				Journey: "account.get",
				Steps: []verification.Step{
					{
						Route: "GET /accounts/{id}/v1",
						Want:  http.StatusOK,
						Got:   http.StatusInternalServerError,
						Exchange: verification.Exchange{
							URL: "http://app/accounts/42/v1",
						},
					},
				},
				Failure: "unexpected status: GET /accounts/{id}/v1 answered 500, want 200",
			},
		},
	}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenTheRequestCantBeSent(t *testing.T) {
	mockResult := _http.MockHTTPResult{Err: errors.New("boom")}
	mockClient, _ := _http.MockSequenceClient(mockResult)
	mockRequest := verification.MockRequest()
	mockJourneys := []verification.Journey{verification.MockJourney("account.get", mockRequest)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{
		Results: []verification.Result{
			{
				Journey: "account.get",
				Steps: []verification.Step{
					{Route: "GET /accounts/{id}/v1", Want: http.StatusOK, Exchange: verification.Exchange{URL: "http://app/accounts/42/v1"}},
				},
				Failure: `can't send the request: failed to execute request: Get "http://app/accounts/42/v1": boom`,
			},
		},
	}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenTheRequestIsInvalid(t *testing.T) {
	mockClient, _ := _http.MockSequenceClient()
	mockRequest := verification.MockRequestWith("/accounts/{id}/v1", http.StatusOK)
	mockJourneys := []verification.Journey{verification.MockJourney("account.get", mockRequest)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{Journey: "account.get", Failure: "invalid request: the route has no method"}}}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenTheBodyIsNotJSON(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: "pong"}
	mockClient, _ := _http.MockSequenceClient(mockResult)
	mockRequest := verification.MockRequest()
	mockTarget := map[string]any{}
	mockJourneys := []verification.Journey{verification.MockDecodingJourney("health.ping", mockRequest, &mockTarget)}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{
		Results: []verification.Result{
			{
				Journey: "health.ping",
				Steps: []verification.Step{
					{
						Route: "GET /accounts/{id}/v1",
						Want:  http.StatusOK,
						Got:   http.StatusOK,
						Exchange: verification.Exchange{
							URL: "http://app/accounts/42/v1",
						},
					},
				},
				Failure: "can't decode the response body: invalid character 'p' looking for beginning of value",
			},
		},
	}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenARequirementIsNotMet(t *testing.T) {
	mockClient, _ := _http.MockSequenceClient()
	mockJourneys := []verification.Journey{verification.MockRequiringJourney("session.revoke", false, "the other session is still active")}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{
		Results: []verification.Result{
			{Journey: "session.revoke", Failure: "requirement not met: the other session is still active"},
		},
	}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenAJourneyPanics(t *testing.T) {
	mockClient, _ := _http.MockSequenceClient()
	mockJourneys := []verification.Journey{
		verification.MockPanickingJourney("session.revoke", "index out of range"),
		verification.MockRequiringJourney("profile.get", true, "unused"),
	}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{
		Results: []verification.Result{
			{Journey: "session.revoke", Failure: "journey panicked: index out of range"},
			{Journey: "profile.get"},
		},
	}
	got := drive(context.Background(), mockJourneys)

	assert.Equal(t, want, got)
}
