package request

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ReadBody reads the whole body of r, which can't be larger than limit bytes. A larger body stops the read and
// returns ErrBodyTooLarge; any other failure returns ErrCantReadBody. Use it for a raw body, such as an uploaded
// file, that Bind does not decode.
func ReadBody(w http.ResponseWriter, r *http.Request, limit int64) ([]byte, error) {
	r.Body = http.MaxBytesReader(w, r.Body, limit)

	content, err := io.ReadAll(r.Body)
	if err != nil {
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			return nil, fmt.Errorf("%w: %w", ErrBodyTooLarge, err)
		}

		return nil, fmt.Errorf("%w: %w", ErrCantReadBody, err)
	}

	return content, nil
}
