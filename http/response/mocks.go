package response

import (
	"fmt"
	"html"
	"net/http"
)

// MockRedirected returns the Recorded that http.Redirect with status 302 writes for a GET to location.
// The cookies are the Set-Cookie values the handler set before redirecting, in order.
func MockRedirected(location string, cookies ...string) Recorded[any] {
	header := http.Header{
		"Content-Type": {"text/html; charset=utf-8"},
		"Location":     {location},
	}
	if len(cookies) > 0 {
		header["Set-Cookie"] = cookies
	}

	return Recorded[any]{
		Status: http.StatusFound,
		Header: header,
		Body:   fmt.Sprintf("<a href=\"%s\">Found</a>.\n\n", html.EscapeString(location)),
	}
}

// MockOK returns the Recorded of a handler that answers 200 with message and data.
func MockOK[T any](message string, data T) Recorded[T] {
	return Recorded[T]{
		Status:  http.StatusOK,
		Message: message,
		Data:    data,
	}
}

// MockCreated returns the Recorded of a handler that answers 201 with message and data.
func MockCreated[T any](message string, data T) Recorded[T] {
	return Recorded[T]{
		Status:  http.StatusCreated,
		Message: message,
		Data:    data,
	}
}

// MockBadRequest returns the Recorded of a handler that answers 400 with message and no data.
func MockBadRequest(message string) Recorded[any] {
	return Recorded[any]{
		Status:  http.StatusBadRequest,
		Message: message,
	}
}

// MockNotFound returns the Recorded of a handler that answers 404 with message and no data.
func MockNotFound(message string) Recorded[any] {
	return Recorded[any]{
		Status:  http.StatusNotFound,
		Message: message,
	}
}

// MockInternalServerError returns the Recorded of a handler that answers 500 with message and no data.
func MockInternalServerError(message string) Recorded[any] {
	return Recorded[any]{
		Status:  http.StatusInternalServerError,
		Message: message,
	}
}
