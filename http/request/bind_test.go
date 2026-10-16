package request_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/request"
)

var errMissingRole = errors.New("role is required")

type roleBody struct {
	Role string `json:"role"`
}

func (b roleBody) IsValid() error {
	if b.Role == "" {
		return errMissingRole
	}

	return nil
}

type pointerBody struct {
	Role string `json:"role"`
}

func (b *pointerBody) IsValid() error {
	if b.Role == "" {
		return errMissingRole
	}

	return nil
}

func TestBind_success(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/accounts/Nebula?limit=20&from=2026-01-02T03:04:05Z", strings.NewReader(`{"role":"administrator"}`))
	req.SetPathValue("username", " Nebula ")
	path := request.Fields{"username": request.String}
	query := request.Fields{"limit": request.Int, "from": request.Time, "actor": request.String.Optional()}

	got, err := request.Bind[roleBody](req, path, query)

	assert.NoError(t, err)
	assert.Equal(t, "Nebula", got.Path.String("username"))
	assert.Equal(t, 20, got.Query.Int("limit"))
	assert.Equal(t, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), got.Query.Time("from"))
	assert.Equal(t, roleBody{Role: "administrator"}, got.Body)
}

func TestBind_successWhenOptionalParamIsLeftOut(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	query := request.Fields{"limit": request.Int.Optional()}

	got, err := request.Bind[request.NoBody](req, nil, query)

	assert.NoError(t, err)
	assert.False(t, got.Query.Has("limit"))
	assert.Equal(t, 0, got.Query.Int("limit"))
}

func TestBind_successWhenOptionalParamIsSentWithItsZeroValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?confirm=false", nil)
	query := request.Fields{"confirm": request.Bool.Optional()}

	got, err := request.Bind[request.NoBody](req, nil, query)

	assert.NoError(t, err)
	assert.True(t, got.Query.Has("confirm"))
	assert.False(t, got.Query.Bool("confirm"))
}

func TestBind_successWhenBodyValidatesWithPointerReceiver(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`{"role":"standard"}`))

	got, err := request.Bind[pointerBody](req, nil, nil)

	assert.NoError(t, err)
	assert.Equal(t, "standard", got.Body.Role)
}

func TestBind_successWhenNoBodyDoesNotReadTheBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/roles", strings.NewReader(`not json`))

	_, err := request.Bind[request.NoBody](req, nil, nil)

	assert.NoError(t, err)
}

func TestBind_failsWhenPathParamIsMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/accounts/", nil)
	req.SetPathValue("username", "   ")
	path := request.Fields{"username": request.String}

	_, err := request.Bind[request.NoBody](req, path, nil)

	want := request.ErrMissingPathParam
	got := err

	assert.ErrorIs(t, got, want)
	assert.Equal(t, "missing path param: username", err.Error())
}

func TestBind_failsWhenPathParamIsInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/sessions/abc", nil)
	req.SetPathValue("id", "abc")
	path := request.Fields{"id": request.Int}

	_, err := request.Bind[request.NoBody](req, path, nil)

	want := request.ErrInvalidPathParam
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenRequiredParamIsItsZeroValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?limit=0", nil)
	query := request.Fields{"limit": request.Int}

	_, err := request.Bind[request.NoBody](req, nil, query)

	want := request.ErrMissingQueryParam
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenQueryParamIsInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?from=yesterday", nil)
	query := request.Fields{"from": request.Time.Optional()}

	_, err := request.Bind[request.NoBody](req, nil, query)

	want := request.ErrInvalidQueryParam
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenBodyIsInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`not json`))

	_, err := request.Bind[roleBody](req, nil, nil)

	want := request.ErrInvalidBody
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenBodyIsNotValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`{}`))

	_, err := request.Bind[roleBody](req, nil, nil)

	want := errMissingRole
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWithTheFirstParamInNameOrder(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/deletions", nil)
	path := request.Fields{"username": request.String, "planet": request.String}

	_, err := request.Bind[request.NoBody](req, path, nil)

	want := "missing path param: planet"
	got := err.Error()

	assert.Equal(t, want, got)
}

func TestValues_panicsWhenParamIsNotDeclaredWithThatKind(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?limit=5", nil)
	query := request.Fields{"limit": request.Int}

	got, err := request.Bind[request.NoBody](req, nil, query)

	assert.NoError(t, err)
	assert.Panics(t, func() { got.Query.String("limit") })
	assert.Panics(t, func() { got.Query.Int("offset") })
}
