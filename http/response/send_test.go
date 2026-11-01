package response_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/response"
	"github.com/lhbelfanti/ditto/v2/log"
)

func TestSend_success(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
		data    any
		err     error
		want    string
	}{
		{name: "WithData", code: http.StatusOK, message: "Request successful", data: map[string]string{"key": "value"}, want: `{"code":200,"message":"Request successful","data":{"key":"value"}}`},
		{name: "WithoutData", code: http.StatusNoContent, message: "No content", want: `{"code":204,"message":"No content"}`},
		{name: "WithAnErrorThatStaysOutOfTheBody", code: http.StatusUnauthorized, message: "Unauthorized access", err: errors.New("secret cause"), want: `{"code":401,"message":"Unauthorized access"}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			response.Send(context.Background(), w, tt.code, tt.message, tt.data, tt.err)

			got := w.Body.String()

			assert.JSONEq(t, tt.want, got)
		})
	}
}

func TestSend_successWhenCodeIsAnErrorWithACause(t *testing.T) {
	mockOutput := log.MockLogOutput(t)
	w := httptest.NewRecorder()

	response.Send(context.Background(), w, http.StatusInternalServerError, "request failed", nil, errors.New("cause"))

	want := []map[string]any{{"level": "error", "message": "request failed", "error": "cause"}}
	got := mockOutput.Entries()

	assert.Equal(t, want, got)
}

func TestSend_successWhenCodeIsAnErrorWithoutACause(t *testing.T) {
	mockOutput := log.MockLogOutput(t)
	w := httptest.NewRecorder()

	response.Send(context.Background(), w, http.StatusNotFound, "not found", nil, nil)

	want := []map[string]any{{"level": "error", "message": "not found"}}
	got := mockOutput.Entries()

	assert.Equal(t, want, got)
}

func TestSend_successWhenCodeIsNotAnError(t *testing.T) {
	mockOutput := log.MockLogOutput(t)
	w := httptest.NewRecorder()

	response.Send(context.Background(), w, http.StatusOK, "done", nil, nil)

	want := []map[string]any{{"level": "info", "message": "done"}}
	got := mockOutput.Entries()

	assert.Equal(t, want, got)
}
