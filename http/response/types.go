package response

import "net/http"

type (
	// Recorded is the whole response a handler wrote, for a test to compare with a whole want. Header has every
	// header but the JSON content type that Send sets, and is nil when nothing else is left. A body that Send
	// wrote, with the status as its code, fills Message and Data; any other body, such as the page of a redirect, stays as it is in Body.
	Recorded[T any] struct {
		Status  int
		Header  http.Header
		Message string
		Data    T
		Body    string
	}

	// Body is the body Send writes, read back with its data decoded as T.
	Body[T any] struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    T      `json:"data"`
	}
)
