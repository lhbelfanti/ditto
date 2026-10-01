package response_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http/response"
	"github.com/lhbelfanti/ditto/v2/log"
)

func TestSend_success(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
		data    interface{}
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
				Error:   "",
			},
		},
		{
			name:    "Error response with an error",
			code:    http.StatusUnauthorized,
			message: "Unauthorized access",
			data:    nil,
			err:     errors.New("invalid token"),
			want: response.DTO{
				Code:    http.StatusUnauthorized,
				Message: "Unauthorized access",
				Data:    nil,
				Error:   errors.New("invalid token").Error(),
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
				Error:   "",
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
			err := json.NewDecoder(resp.Body).Decode(&got)
			assert.NoError(t, err, "Failed to decode response")

			jsonWant, _ := json.Marshal(tt.want)
			jsonGot, _ := json.Marshal(got)
			assert.JSONEq(t, string(jsonWant), string(jsonGot), "Response mismatch")
		})
	}
}

func TestSend_logsErrorOnce(t *testing.T) {
	var output bytes.Buffer
	log.NewCustomLogger(&output, zerolog.TraceLevel)
	t.Cleanup(func() { log.NewCustomLogger(os.Stdout, zerolog.DebugLevel) })

	w := httptest.NewRecorder()
	response.Send(context.Background(), w, http.StatusInternalServerError, "request failed", nil, errors.New("cause"))

	assert.Equal(t, 1, bytes.Count(output.Bytes(), []byte(`"level":"error"`)))
	assert.Contains(t, output.String(), `"message":"request failed"`)
	assert.Contains(t, output.String(), `"error":"cause"`)
}
