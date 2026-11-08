package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/response"
)

func TestPingHandlerV1_success(t *testing.T) {
	handler := pingHandlerV1()
	req := httptest.NewRequest(http.MethodGet, "/ping/v1", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	want := response.Recorded[any]{Status: http.StatusOK, Message: "pong"}
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestDatabasePingHandlerV1_success(t *testing.T) {
	mockPing := MockDatabasePing(nil)
	handler := databasePingHandlerV1(mockPing)
	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	want := response.Recorded[any]{Status: http.StatusOK, Message: "pong"}
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}

func TestDatabasePingHandlerV1_failsWhenPingFails(t *testing.T) {
	mockPing := MockDatabasePing(errors.New("database unreachable"))
	handler := databasePingHandlerV1(mockPing)
	req := httptest.NewRequest(http.MethodGet, "/database/ping/v1", nil)
	w := httptest.NewRecorder()

	handler(w, req)

	want := response.Recorded[any]{Status: http.StatusServiceUnavailable, Message: ErrMsgDatabaseUnavailable}
	got := response.Read[any](w)

	assert.Equal(t, want, got)
}
