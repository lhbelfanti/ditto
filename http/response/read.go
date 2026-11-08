package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

// contentType is the content type Send writes.
const contentType string = "application/json"

// Read returns the whole response the handler wrote to w, with the data of its body decoded as T.
func Read[T any](w *httptest.ResponseRecorder) Recorded[T] {
	header := w.Header().Clone()
	if header.Get("Content-Type") != contentType {
		return Recorded[T]{Status: w.Code, Header: nilWhenEmpty(header), Body: w.Body.String()}
	}

	var body sentBody[T]
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil || body.Code != w.Code {
		return Recorded[T]{Status: w.Code, Header: nilWhenEmpty(header), Body: w.Body.String()}
	}

	header.Del("Content-Type")

	return Recorded[T]{Status: w.Code, Header: nilWhenEmpty(header), Message: body.Message, Data: body.Data}
}

// nilWhenEmpty returns nil for a header without values, so that a want can leave Header out.
func nilWhenEmpty(header http.Header) http.Header {
	if len(header) == 0 {
		return nil
	}

	return header
}
