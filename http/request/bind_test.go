package request_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/request"
)

func TestBind_success(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/accounts/Nebula?limit=20&from=2026-01-02T03:04:05Z&done=true", strings.NewReader(`{"role":"administrator"}`))
	req.SetPathValue("username", " Nebula ")

	want := request.MockRequestDTO{
		Path:  request.MockAccountPath{Username: "Nebula"},
		Query: request.MockEventsQuery{Limit: 20, From: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), Done: true},
		Body:  request.MockBody{Role: "administrator"},
	}
	got, _ := request.Bind[request.MockRequestDTO](req)

	assert.Equal(t, want, got)
}

func TestBind_successWhenOptionalParamsAreLeftOut(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events", nil)

	want := request.MockOptionalQueryDTO{}
	got, _ := request.Bind[request.MockOptionalQueryDTO](req)

	assert.Equal(t, want, got)
}

func TestBind_successWhenOptionalParamsAreSentWithTheirZeroValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?limit=0&confirm=false", nil)
	limit := 0
	confirm := false

	want := request.MockOptionalQueryDTO{Query: request.MockOptionalQuery{Limit: &limit, Confirm: &confirm}}
	got, _ := request.Bind[request.MockOptionalQueryDTO](req)

	assert.Equal(t, want, got)
}

func TestBind_successWhenTheDTOHasOnlyPath(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/sessions/42?ignored=1", strings.NewReader(`not json`))
	req.SetPathValue("id", "42")

	want := request.MockSessionDTO{Path: request.MockSessionPath{ID: 42}}
	got, _ := request.Bind[request.MockSessionDTO](req)

	assert.Equal(t, want, got)
}

func TestBind_successWhenTheDTOHasNoFields(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health?ignored=1", strings.NewReader(`not json`))

	_, got := request.Bind[struct{}](req)

	assert.NoError(t, got)
}

func TestBind_failsWhenTheDTOIsNotAStruct(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	_, got := request.Bind[string](req)
	want := request.ErrInvalidDTO

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenPathIsInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/accounts/", nil)
	req.SetPathValue("username", "   ")

	_, got := request.Bind[request.MockRequestDTO](req)
	want := request.ErrInvalidPath

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenQueryIsInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?confirm=maybe", nil)

	_, got := request.Bind[request.MockOptionalQueryDTO](req)
	want := request.ErrInvalidQuery

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenBodyDoesNotDecode(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/accounts/Nebula?limit=20&from=2026-01-02T03:04:05Z&done=true", strings.NewReader(`not json`))
	req.SetPathValue("username", "Nebula")

	_, got := request.Bind[request.MockRequestDTO](req)
	want := request.ErrInvalidBody

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenBodyIsNotValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`{}`))

	_, got := request.Bind[request.MockInvalidBodyDTO](req)
	want := request.ErrMockInvalidBody

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenBodyWithPointerReceiverIsNotValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`{}`))

	_, got := request.Bind[request.MockInvalidPointerBodyDTO](req)
	want := request.ErrMockInvalidBody

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWithTheFirstMissingParamInFieldOrder(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/deletions", nil)

	_, err := request.Bind[request.MockDeletionDTO](req)

	want := "invalid path: missing param: username"
	got := err.Error()

	assert.Equal(t, want, got)
}
