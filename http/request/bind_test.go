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

type changeRoleDTO struct {
	Path struct {
		Username string `path:"username"`
	}
	Query struct {
		Limit int       `query:"limit"`
		From  time.Time `query:"from"`
		Actor *string   `query:"actor"`
	}
	Body roleBody
}

type eventsDTO struct {
	Query struct {
		Limit   *int  `query:"limit"`
		Confirm *bool `query:"confirm"`
	}
}

type sessionDTO struct {
	Path struct {
		ID int64 `path:"id"`
	}
}

type deletionDTO struct {
	Path struct {
		Username string `path:"username"`
		Planet   string `path:"planet"`
	}
}

func TestBind_success(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/accounts/Nebula?limit=20&from=2026-01-02T03:04:05Z", strings.NewReader(`{"role":"administrator"}`))
	req.SetPathValue("username", " Nebula ")

	got, err := request.Bind[changeRoleDTO](req)

	assert.NoError(t, err)
	assert.Equal(t, "Nebula", got.Path.Username)
	assert.Equal(t, 20, got.Query.Limit)
	assert.Equal(t, time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), got.Query.From)
	assert.Nil(t, got.Query.Actor)
	assert.Equal(t, roleBody{Role: "administrator"}, got.Body)
}

func TestBind_successWhenOptionalParamIsLeftOut(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events", nil)

	got, err := request.Bind[eventsDTO](req)

	assert.NoError(t, err)
	assert.Nil(t, got.Query.Limit)
	assert.Nil(t, got.Query.Confirm)
}

func TestBind_successWhenOptionalParamIsSentWithItsZeroValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?limit=0&confirm=false", nil)

	got, err := request.Bind[eventsDTO](req)

	assert.NoError(t, err)
	assert.Equal(t, 0, *got.Query.Limit)
	assert.False(t, *got.Query.Confirm)
}

func TestBind_successWhenBodyValidatesWithPointerReceiver(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`{"role":"standard"}`))

	got, err := request.Bind[struct{ Body pointerBody }](req)

	assert.NoError(t, err)
	assert.Equal(t, "standard", got.Body.Role)
}

func TestBind_successWhenTheDTOHasNoBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events", strings.NewReader(`not json`))

	_, err := request.Bind[eventsDTO](req)

	assert.NoError(t, err)
}

func TestBind_failsWhenPathParamIsMissing(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/accounts/", nil)
	req.SetPathValue("username", "   ")

	_, err := request.Bind[changeRoleDTO](req)

	want := request.ErrMissingPathParam
	got := err

	assert.ErrorIs(t, got, want)
	assert.Equal(t, "missing path param: username", err.Error())
}

func TestBind_failsWhenPathParamIsInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/sessions/abc", nil)
	req.SetPathValue("id", "abc")

	_, err := request.Bind[sessionDTO](req)

	want := request.ErrInvalidPathParam
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenRequiredParamIsItsZeroValue(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/accounts/Nebula?limit=0", nil)
	req.SetPathValue("username", "Nebula")

	_, err := request.Bind[changeRoleDTO](req)

	want := request.ErrMissingQueryParam
	got := err

	assert.ErrorIs(t, got, want)
	assert.Equal(t, "missing query param: limit", err.Error())
}

func TestBind_failsWhenQueryParamIsInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/events?confirm=maybe", nil)

	_, err := request.Bind[eventsDTO](req)

	want := request.ErrInvalidQueryParam
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenBodyIsInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`not json`))

	_, err := request.Bind[struct{ Body roleBody }](req)

	want := request.ErrInvalidBody
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWhenBodyIsNotValid(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/roles", strings.NewReader(`{}`))

	_, err := request.Bind[struct{ Body roleBody }](req)

	want := errMissingRole
	got := err

	assert.ErrorIs(t, got, want)
}

func TestBind_failsWithTheFirstParamInFieldOrder(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/deletions", nil)

	_, err := request.Bind[deletionDTO](req)

	want := "missing path param: username"
	got := err.Error()

	assert.Equal(t, want, got)
}

func TestBind_failsWhenTheDTOIsInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/accounts/Nebula?ids=1", nil)
	req.SetPathValue("username", "Nebula")

	_, errNotStruct := request.Bind[string](req)
	_, errUntagged := request.Bind[struct{ Path struct{ Username string } }](req)
	_, errUnsupported := request.Bind[struct {
		Query struct {
			IDs []int `query:"ids"`
		}
	}](req)

	assert.ErrorIs(t, errNotStruct, request.ErrInvalidDTO)
	assert.ErrorIs(t, errUntagged, request.ErrInvalidDTO)
	assert.ErrorIs(t, errUnsupported, request.ErrInvalidDTO)
}
