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
	mockStep := verification.Step{Description: "reads the account", Method: http.MethodGet, Path: "/accounts/42/v1", Status: http.StatusOK}
	mockEntries := []verification.Entry{{ID: "account.get", Journey: verification.MockJourney(mockStep)}}

	drive := verification.MakeDrive(mockClient, "http://app/")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.get",
		Records: []verification.Record{{
			Description: "reads the account",
			Want:        http.StatusOK,
			Got:         http.StatusOK,
			Exchange: verification.Exchange{
				Method:         http.MethodGet,
				URL:            "http://app/accounts/42/v1",
				ResponseHeader: map[string]string{"Content-Type": "application/json"},
				ResponseBody:   map[string]any{"id": float64(42), "token": "<redacted>"},
			},
		}},
	}}}
	got := drive(context.Background(), mockEntries)

	assert.Equal(t, want, got)
}

func TestDrive_successWhenTheStepHasPathValuesAQueryABearerABodyAndAHeader(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK}
	mockClient := verification.MockClient(mockResult)
	mockPathValues := map[string]string{"id": "42"}
	mockStep := verification.Step{
		Description: "updates the account",
		Method:      http.MethodPut,
		Path:        "/accounts/{id}/v1",
		PathValues:  verification.MockValues(mockPathValues),
		Query:       map[string]string{"notify": "true"},
		Bearer:      "session",
		Body:        map[string]string{"name": "Ana", "password": "pw"},
		With:        verification.MockWithHeader("Accept", "application/json"),
		Status:      http.StatusOK,
	}
	mockEntries := []verification.Entry{{ID: "account.update", Journey: verification.MockJourney(mockStep)}}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.update",
		Records: []verification.Record{{
			Description: "updates the account",
			Want:        http.StatusOK,
			Got:         http.StatusOK,
			Exchange: verification.Exchange{
				Method:        http.MethodPut,
				URL:           "http://app/accounts/42/v1?notify=true",
				RequestHeader: map[string]string{"Accept": "application/json", "Authorization": "<redacted>", "Content-Type": "application/json"},
				RequestBody:   map[string]any{"name": "Ana", "password": "<redacted>"},
			},
		}},
	}}}
	got := drive(context.Background(), mockEntries)

	assert.Equal(t, want, got)
}

func TestDrive_successWhenTheJourneySetsItselfUp(t *testing.T) {
	mockHeader := http.Header{"Location": {"http://google-mock/authorize?state=abc"}}
	mockSetupResult := _http.MockHTTPResult{StatusCode: http.StatusFound, Header: mockHeader}
	mockStepResult := _http.MockHTTPResult{StatusCode: http.StatusOK}
	mockClient := verification.MockClient(mockSetupResult, mockStepResult)
	mockStep := verification.Step{Description: "pings", Method: http.MethodGet, Path: "/ping/v1", Status: http.StatusOK}
	mockEntries := []verification.Entry{{ID: "signin", Journey: verification.MockSettingUpJourney("/auth/google/v1", http.StatusFound, mockStep)}}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "signin",
		Records: []verification.Record{
			{
				Want:  http.StatusFound,
				Got:   http.StatusFound,
				Setup: true,
				Exchange: verification.Exchange{
					Method:         http.MethodGet,
					URL:            "http://app/auth/google/v1",
					ResponseHeader: map[string]string{"Location": "http://google-mock/authorize?state=%3Credacted%3E"},
				},
			},
			{
				Description: "pings",
				Want:        http.StatusOK,
				Got:         http.StatusOK,
				Exchange:    verification.Exchange{Method: http.MethodGet, URL: "http://app/ping/v1"},
			},
		},
	}}}
	got := drive(context.Background(), mockEntries)

	assert.Equal(t, want, got)
}

func TestDrive_successWhenTheStepAnswersTheExpectedData(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: `{"code":200,"message":"ok","data":{"id":42}}`}
	mockClient := verification.MockClient(mockResult)
	mockStep := verification.Step{
		Description: "reads the account",
		Method:      http.MethodGet,
		Path:        "/accounts/42/v1",
		Status:      http.StatusOK,
		Expect:      verification.Data(map[string]int{"id": 42}),
	}
	mockEntries := []verification.Entry{{ID: "account.get", Journey: verification.MockJourney(mockStep)}}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.get",
		Records: []verification.Record{{
			Description: "reads the account",
			Want:        http.StatusOK,
			Got:         http.StatusOK,
			Exchange: verification.Exchange{
				Method:       http.MethodGet,
				URL:          "http://app/accounts/42/v1",
				ResponseBody: map[string]any{"code": float64(200), "message": "ok", "data": map[string]any{"id": float64(42)}},
			},
		}},
	}}}
	got := drive(context.Background(), mockEntries)

	assert.Equal(t, want, got)
}

func TestDrive_successWhenTheJourneyAfterAFailedOneRuns(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK}
	mockClient := verification.MockClient(mockResult)
	mockErr := errors.New("boom")
	mockStep := verification.Step{Description: "pings", Method: http.MethodGet, Path: "http://other/ping/v1", Status: http.StatusOK}
	mockEntries := []verification.Entry{
		{ID: "session.revoke", Journey: verification.MockFailingJourney(mockErr)},
		{ID: "health.ping", Journey: verification.MockJourney(mockStep)},
	}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{
		{Journey: "session.revoke", Failure: mockErr},
		{
			Journey: "health.ping",
			Records: []verification.Record{{
				Description: "pings",
				Want:        http.StatusOK,
				Got:         http.StatusOK,
				Exchange:    verification.Exchange{Method: http.MethodGet, URL: "http://other/ping/v1"},
			}},
		},
	}}
	got := drive(context.Background(), mockEntries)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenAStepAnswersAnotherStatus(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusNotFound}
	mockClient := verification.MockClient(mockResult)
	mockStep := verification.Step{Description: "reads the account", Method: http.MethodGet, Path: "/accounts/42/v1", Status: http.StatusOK}
	mockEntries := []verification.Entry{{ID: "account.get", Journey: verification.MockJourney(mockStep, mockStep)}}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.get",
		Records: []verification.Record{{
			Description: "reads the account",
			Want:        http.StatusOK,
			Got:         http.StatusNotFound,
			Exchange:    verification.Exchange{Method: http.MethodGet, URL: "http://app/accounts/42/v1"},
		}},
		Failure: fmt.Errorf("%w: GET http://app/accounts/42/v1 answered 404, want 200", verification.ErrUnexpectedStatus),
	}}}
	got := drive(context.Background(), mockEntries)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenTheRequestCantBeSent(t *testing.T) {
	mockErr := errors.New("boom")
	mockResult := _http.MockHTTPResult{Err: mockErr}
	mockClient := verification.MockClient(mockResult)
	mockStep := verification.Step{Description: "reads the account", Method: http.MethodGet, Path: "/accounts/42/v1", Status: http.StatusOK}
	mockEntries := []verification.Entry{{ID: "account.get", Journey: verification.MockJourney(mockStep)}}
	mockURLErr := &url.Error{Op: "Get", URL: "http://app/accounts/42/v1", Err: mockErr}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.get",
		Records: []verification.Record{{
			Description: "reads the account",
			Want:        http.StatusOK,
			Exchange:    verification.Exchange{Method: http.MethodGet, URL: "http://app/accounts/42/v1"},
		}},
		Failure: fmt.Errorf("%w: %w", verification.ErrCantSendRequest, fmt.Errorf("%w: %w", _http.ErrFailedToExecuteRequest, mockURLErr)),
	}}}
	got := drive(context.Background(), mockEntries)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenTheStepAnswersOtherData(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: `{"code":200,"data":{"id":7}}`}
	mockClient := verification.MockClient(mockResult)
	mockStep := verification.Step{
		Description: "reads the account",
		Method:      http.MethodGet,
		Path:        "/accounts/42/v1",
		Status:      http.StatusOK,
		Expect:      verification.Data(map[string]int{"id": 42}),
	}
	mockEntries := []verification.Entry{{ID: "account.get", Journey: verification.MockJourney(mockStep)}}
	mockDataErr := fmt.Errorf("%w: got %+v, want %+v", verification.ErrUnexpectedData, map[string]int{"id": 7}, map[string]int{"id": 42})

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "account.get",
		Records: []verification.Record{{
			Description: "reads the account",
			Want:        http.StatusOK,
			Got:         http.StatusOK,
			Exchange: verification.Exchange{
				Method:       http.MethodGet,
				URL:          "http://app/accounts/42/v1",
				ResponseBody: map[string]any{"code": float64(200), "data": map[string]any{"id": float64(7)}},
			},
		}},
		Failure: fmt.Errorf("%w: %s: %w", verification.ErrUnexpectedResponse, "reads the account", mockDataErr),
	}}}
	got := drive(context.Background(), mockEntries)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenTheSetupCantDecodeTheBody(t *testing.T) {
	mockResult := _http.MockHTTPResult{StatusCode: http.StatusOK, Body: "pong"}
	mockClient := verification.MockClient(mockResult)
	var mockTarget map[string]any
	mockEntries := []verification.Entry{{ID: "health.ping", Journey: verification.MockDecodingJourney("/ping/v1", &mockTarget)}}
	mockDecodeErr := verification.MockDecodeError("pong")

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "health.ping",
		Records: []verification.Record{{
			Want:     http.StatusOK,
			Got:      http.StatusOK,
			Setup:    true,
			Exchange: verification.Exchange{Method: http.MethodGet, URL: "http://app/ping/v1"},
		}},
		Failure: fmt.Errorf("%w: %w", verification.ErrCantDecodeBody, mockDecodeErr),
	}}}
	got := drive(context.Background(), mockEntries)

	assert.Equal(t, want, got)
}

func TestDrive_failsWhenAJourneyPanics(t *testing.T) {
	mockClient := verification.MockClient()
	mockEntries := []verification.Entry{{ID: "session.list", Journey: verification.MockPanickingJourney("index out of range")}}

	drive := verification.MakeDrive(mockClient, "http://app")

	want := verification.Report{Results: []verification.Result{{
		Journey: "session.list",
		Failure: fmt.Errorf("%w: %v", verification.ErrJourneyPanicked, "index out of range"),
	}}}
	got := drive(context.Background(), mockEntries)

	assert.Equal(t, want, got)
}
