package response_test

import (
	"bytes"
	"context"
	"encoding/json"
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
		want    response.DTO
	}{
		{
			name:    "Success response with data",
			code:    http.StatusOK,
			message: "Request successful",
			data:    map[string]string{"key": "value"},
			err:     nil,
			want: response.DTO{
				Code:    http.StatusOK,
				Message: "Request successful",
				Data:    map[string]string{"key": "value"},
			},
		},
		{
			name:    "Error response with an error hides the cause",
			code:    http.StatusUnauthorized,
			message: "Unauthorized access",
			data:    nil,
			err:     errors.New("invalid token"),
			want: response.DTO{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized access",
				Data:    nil,
			},
		},
		{
			name:    "Empty data response",
			code:    http.StatusNoContent,
			message: "No content",
			data:    nil,
			err:     nil,
			want: response.DTO{
				Code:    http.StatusNoContent,
				Message: "No content",
				Data:    nil,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()

			response.Send(context.Background(), w, tt.code, tt.message, tt.data, tt.err)
			resp := w.Result()
			defer resp.Body.Close()

			var got response.DTO
			_ = json.NewDecoder(resp.Body).Decode(&got)

			jsonWant, _ := json.Marshal(tt.want)
			jsonGot, _ := json.Marshal(got)

			assert.JSONEq(t, string(jsonWant), string(jsonGot))
		})
	}
}

func TestSend_successWhenErrorIsLoggedOnce(t *testing.T) {
	output := log.MockLogOutput(t)
	w := httptest.NewRecorder()

	response.Send(context.Background(), w, http.StatusInternalServerError, "request failed", nil, errors.New("cause"))

	want := 1
	got := bytes.Count(output.Bytes(), []byte(`"level":"error"`))

	assert.Equal(t, want, got)
}

func TestSend_successWhenErrorIsLoggedWithMessage(t *testing.T) {
	output := log.MockLogOutput(t)
	w := httptest.NewRecorder()

	response.Send(context.Background(), w, http.StatusInternalServerError, "request failed", nil, errors.New("cause"))

	assert.Contains(t, output.String(), `"message":"request failed"`)
}

func TestSend_successWhenErrorIsLoggedWithCause(t *testing.T) {
	output := log.MockLogOutput(t)
	w := httptest.NewRecorder()

	response.Send(context.Background(), w, http.StatusInternalServerError, "request failed", nil, errors.New("cause"))

	assert.Contains(t, output.String(), `"error":"cause"`)
}

func TestSend_successWhenCauseIsNotInTheResponseBody(t *testing.T) {
	w := httptest.NewRecorder()

	response.Send(context.Background(), w, http.StatusInternalServerError, "request failed", nil, errors.New("secret cause"))

	assert.NotContains(t, w.Body.String(), "secret cause")
}
