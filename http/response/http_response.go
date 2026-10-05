package response

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/lhbelfanti/ditto/v2/log"
)

// Send writes a standardized JSON response to the client.
// It accepts an HTTP status code, a message, optional data, and optional error details.
// The response format includes the code, message, and either data or error information.
func Send(ctx context.Context, w http.ResponseWriter, code int, message string, data any, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	resp := DTO{
		Code:    code,
		Message: message,
		Data:    data,
	}

	if code >= 400 {
		if err != nil {
			resp.Error = err.Error()
			log.Err(ctx, err, message)
		} else {
			log.Error(ctx, message)
		}
	} else {
		log.Info(ctx, message)
	}

	_ = json.NewEncoder(w).Encode(resp)
}
