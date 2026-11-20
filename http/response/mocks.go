package response

import (
	"fmt"
	"html"
	"net/http"
)

// MockRedirected returns the Recorded that http.Redirect writes for a GET with status 302 to location.
// The cookies are the Set-Cookie values the handler set before redirecting, in order.
func MockRedirected(location string, cookies ...string) Recorded[any] {
	return MockRedirectedWith(http.StatusFound, location, cookies...)
}

// MockRedirectedWith returns the Recorded that http.Redirect writes for a GET with the given redirect status
// (301, 302, 303, 307 or 308) to location. The page of the body names the status, as http.Redirect does.
func MockRedirectedWith(status int, location string, cookies ...string) Recorded[any] {
	header := http.Header{
		"Content-Type": {"text/html; charset=utf-8"},
		"Location":     {location},
	}
	if len(cookies) > 0 {
		header["Set-Cookie"] = cookies
	}

	return Recorded[any]{
		Status: status,
		Header: header,
		Body:   fmt.Sprintf("<a href=\"%s\">%s</a>.\n\n", html.EscapeString(location), http.StatusText(status)),
	}
}
