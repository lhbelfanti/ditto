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
	var body roleBody

	got, err := request.New(req).WithPath(path).WithQuery(query).WithBody(&body).Bind()

	assert.NoError(t, err)
	assert.Equal(t, roleBody{Role: "administrator"}, body)

	username, err := got.Path.String("username")
	assert.NoError(t, err)
	assert.Equal(t, "Nebula", username)

	limit, err := got.Query.Int("limit")
	assert.NoError(t, err)
	assert.Equal(t, 20, limit)

	from, err := got.Query.Time("from")
	assert.NoError(t, err)
	assert.Equal(t, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), from)
}

func TestBind_successWhenOptionalParamIsLeftOut(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	query := request.Fields{"limit": request.Int.Optional()}

	got, err := request.New(req).WithQuery(query).Bind()

	assert.NoError(t, err)
	assert.False(t, got.Query.Has("limit"))

	limit, err := got.Query.Int("limit")
	assert.NoError(t, err)
	assert.Equal(t, 0, limit)
}

func TestBind_successWhenOptionalParamIsSentWithItsZeroValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?confirm=false", nil)
	query := request.Fields{"confirm": request.Bool.Optional()}

	got, err := request.New(req).WithQuery(query).Bind()

	assert.NoError(t, err)
	assert.True(t, got.Query.Has("confirm"))

	confirm, err := got.Query.Bool("confirm")
	assert.NoError(t, err)
	assert.False(t, confirm)
}

func TestBind_successWhenBodyValidatesWithPointerReceiver(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`{"role":"standard"}`))
	var body pointerBody

	_, err := request.New(req).WithBody(&body).Bind()

	assert.NoError(t, err)
	assert.Equal(t, "standard", body.Role)
}

func TestBind_successWhenWithBodyIsNotCalled(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/roles", strings.NewReader(`not json`))

	_, err := request.New(req).Bind()

	assert.NoError(t, err)
}

func TestBind_failsWhenPathParamIsMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/accounts/", nil)
	req.SetPathValue("username", "   ")
	path := request.Fields{"username": request.String}

	_, err := request.New(req).WithPath(path).Bind()

	want := request.ErrMissingPathParam
	got := err

	assert.ErrorIs(t, got, want)
	assert.Equal(t, "missing path param: username", err.Error())
}

func TestBind_failsWhenPathParamIsInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/sessions/abc", nil)
	req.SetPathValue("id", "abc")
	path := request.Fields{"id": request.Int}

	_, err := request.New(req).WithPath(path).Bind()

	want := request.ErrInvalidPathParam
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenRequiredParamIsItsZeroValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?limit=0", nil)
	query := request.Fields{"limit": request.Int}

	_, err := request.New(req).WithQuery(query).Bind()

	want := request.ErrMissingQueryParam
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenQueryParamIsInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?from=yesterday", nil)
	query := request.Fields{"from": request.Time.Optional()}

	_, err := request.New(req).WithQuery(query).Bind()

	want := request.ErrInvalidQueryParam
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenBodyIsInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`not json`))
	var body roleBody

	_, err := request.New(req).WithBody(&body).Bind()

	want := request.ErrInvalidBody
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenBodyIsNotValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`{}`))
	var body roleBody

	_, err := request.New(req).WithBody(&body).Bind()

	want := errMissingRole
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWithTheFirstParamInNameOrder(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/deletions", nil)
	path := request.Fields{"username": request.String, "planet": request.String}

	_, err := request.New(req).WithPath(path).Bind()

	want := "missing path param: planet"
	got := err.Error()

	assert.Equal(t, want, got)
}

func TestValues_failsWhenQueryParamIsNotDeclared(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?limit=5", nil)
	query := request.Fields{"limit": request.Int}

	got, err := request.New(req).WithQuery(query).Bind()
	assert.NoError(t, err)

	_, err = got.Query.Int("offset")

	assert.ErrorIs(t, err, request.ErrUndeclaredQueryParam)
	assert.Equal(t, "undeclared query param: offset as Int", err.Error())
}

func TestValues_failsWhenQueryParamIsDeclaredWithAnotherKind(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?limit=5", nil)
	query := request.Fields{"limit": request.Int}

	got, err := request.New(req).WithQuery(query).Bind()
	assert.NoError(t, err)

	_, err = got.Query.String("limit")

	want := request.ErrUndeclaredQueryParam
	assert.ErrorIs(t, err, want)
}

func TestValues_failsWhenPathParamIsNotDeclared(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/accounts", nil)

	got, err := request.New(req).Bind()
	assert.NoError(t, err)

	_, err = got.Path.String("username")

	want := request.ErrUndeclaredPathParam
	assert.ErrorIs(t, err, want)
}
