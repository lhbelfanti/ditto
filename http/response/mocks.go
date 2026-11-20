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
