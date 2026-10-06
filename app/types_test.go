package app_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/app"
)

func TestOptions_WithMiddleware_successWhenFirstAddedRunsFirst(t *testing.T) {
	opts := app.Options{Mux: app.MockMux("GET /a", http.StatusOK)}.
		WithMiddleware(app.MockMiddleware("first")).
		WithMiddleware(app.MockMiddleware("second"))
	service := app.MockRunningServiceWith(t, opts)

	want := []string{"first", "second"}
	got := service.Get("/a").Header.Values("X-Order")

	assert.Equal(t, want, got)
}

func TestOptions_WithMiddleware_successWhenOriginalIsNotModified(t *testing.T) {
	base := app.Options{Mux: app.MockMux("GET /a", http.StatusOK)}.WithMiddleware(app.MockMiddleware("base"))
	_ = base.WithMiddleware(app.MockMiddleware("extra"))
	service := app.MockRunningServiceWith(t, base)

	want := []string{"base"}
	got := service.Get("/a").Header.Values("X-Order")

	assert.Equal(t, want, got)
}
