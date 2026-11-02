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

	want := []string{"first", "second"}
	got := service.Get("/a").Header.Values("X-Order")

	assert.Equal(t, want, got)
}

func TestOptions_WithMiddleware_successWhenOriginalIsNotModified(t *testing.T) {
	mockMux := app.MockMux("GET /a", http.StatusOK)
	mockBase := app.MockMiddleware("base")
	mockExtra := app.MockMiddleware("extra")

	base := app.Options{Mux: mockMux}.WithMiddleware(mockBase)
	_ = base.WithMiddleware(mockExtra)
	service := app.StartServiceWith(t, base)

	want := []string{"base"}
	got := service.Get("/a").Header.Values("X-Order")

	assert.Equal(t, want, got)
}
