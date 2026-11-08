package verification

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResolve_success(t *testing.T) {
	request := Request{
		Route: "DELETE /accounts/{id}/sessions/{session}/v1",
		Params: map[string]string{
			"id":      "42",
			"session": "a b",
		},
		Query: url.Values{
			"confirm": {"true"},
		},
	}

	method, target, err := resolve(request, "http://app/")

	want := []any{"DELETE", "http://app/accounts/42/sessions/a%20b/v1?confirm=true", nil}
	got := []any{method, target, err}

	assert.Equal(t, want, got)
}

func TestResolve_successWhenTheTargetIsAnotherService(t *testing.T) {
	request := Request{Route: "GET http://mock/authorize?client_id=neptune", Query: url.Values{"state": {"abc"}}}

	method, target, err := resolve(request, "http://app")

	want := []any{"GET", "http://mock/authorize?client_id=neptune&state=abc", nil}
	got := []any{method, target, err}

	assert.Equal(t, want, got)
}

func TestResolve_failsWhenTheRouteHasNoMethod(t *testing.T) {
	request := Request{Route: "/ping/v1"}

	want := ErrRouteWithoutMethod
	_, _, got := resolve(request, "http://app")

	assert.ErrorIs(t, got, want)
}

func TestResolve_failsWhenAParamHasNoPlaceholder(t *testing.T) {
	request := Request{Route: "GET /ping/v1", Params: map[string]string{"id": "42"}}

	want := ErrUnknownParam
	_, _, got := resolve(request, "http://app")

	assert.ErrorIs(t, got, want)
}

func TestResolve_failsWhenAPlaceholderHasNoParam(t *testing.T) {
	request := Request{Route: "GET /accounts/{id}/v1"}

	want := ErrMissingParam
	_, _, got := resolve(request, "http://app")

	assert.ErrorIs(t, got, want)
}
