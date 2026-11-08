package app_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/app"
)

func TestOptions_WithMiddleware_successWhenCalledTwice(t *testing.T) {
	mockMux := app.MockMux("GET /a", http.StatusOK)
	mockFirst := app.MockMiddleware("first")
	mockSecond := app.MockMiddleware("second")

	opts := app.Options{Mux: mockMux}.WithMiddleware(mockFirst).WithMiddleware(mockSecond)
	service := app.StartServiceWith(t, opts)

	want := app.Reply{Status: http.StatusOK, Header: http.Header{"X-Order": {"first", "second"}}}
	got := service.Get("/a")

	assert.Equal(t, want, got)
}

func TestOptions_WithMiddleware_successWhenOriginalIsNotModified(t *testing.T) {
	mockMux := app.MockMux("GET /a", http.StatusOK)
	mockBase := app.MockMiddleware("base")
	mockExtra := app.MockMiddleware("extra")

	base := app.Options{Mux: mockMux}.WithMiddleware(mockBase)
	_ = base.WithMiddleware(mockExtra)
	service := app.StartServiceWith(t, base)

	want := app.Reply{Status: http.StatusOK, Header: http.Header{"X-Order": {"base"}}}
	got := service.Get("/a")

	assert.Equal(t, want, got)
}

func TestOptions_WithRouteMiddleware_success(t *testing.T) {
	mockMux := app.MockMux("GET /a", http.StatusOK)
	mockRoute := app.MockMiddleware("route")

	opts := app.Options{Mux: mockMux}.WithRouteMiddleware(mockRoute)
	service := app.StartServiceWith(t, opts)

	want := app.Reply{Status: http.StatusOK, Header: http.Header{"X-Order": {"route"}}}
	got := service.Get("/a")

	assert.Equal(t, want, got)
}

func TestOptions_WithRouteMiddleware_successWhenNoRouteMatches(t *testing.T) {
	mockMux := app.MockMux("GET /a", http.StatusOK)
	mockRoute := app.MockMiddleware("route")

	opts := app.Options{Mux: mockMux}.WithRouteMiddleware(mockRoute)
	service := app.StartServiceWith(t, opts)

	want := app.Reply{
		Status: http.StatusNotFound,
		Header: http.Header{
			"Content-Type":           {"text/plain; charset=utf-8"},
			"X-Content-Type-Options": {"nosniff"},
			"X-Order":                {"route"},
		},
		Body: "404 page not found\n",
	}
	got := service.Get("/missing")

	assert.Equal(t, want, got)
}

func TestOptions_WithRouteMiddleware_successWhenRouteIsASystemRoute(t *testing.T) {
	mockMux := app.MockMux("GET /a", http.StatusOK)
	mockRoute := app.MockMiddleware("route")

	opts := app.Options{Mux: mockMux}.WithRouteMiddleware(mockRoute)
	service := app.StartServiceWith(t, opts)

	want := app.Reply{Status: http.StatusOK, Header: http.Header{"Content-Type": {"application/json"}}, Body: `{"code":200,"message":"pong"}` + "\n"}
	got := service.Get("/ping/v1")

	assert.Equal(t, want, got)
}

func TestOptions_WithRouteMiddleware_successWhenWithMiddlewareIsAlsoAdded(t *testing.T) {
	mockMux := app.MockMux("GET /a", http.StatusOK)
	mockRoute := app.MockMiddleware("route")
	mockWhole := app.MockMiddleware("whole")

	opts := app.Options{Mux: mockMux}.WithRouteMiddleware(mockRoute).WithMiddleware(mockWhole)
	service := app.StartServiceWith(t, opts)

	want := app.Reply{Status: http.StatusOK, Header: http.Header{"X-Order": {"whole", "route"}}}
	got := service.Get("/a")

	assert.Equal(t, want, got)
}
