package verification

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	_http "github.com/lhbelfanti/ditto/v2/http"
)

func TestRedact_success(t *testing.T) {
	request := Request{Token: "session", Header: map[string]string{"Cookie": "state=abc"}, Body: map[string]string{"name": "Ana", "password": "pw"}}
	response := _http.Response{Header: http.Header{"Location": {"/done?code=xyz"}}, Body: "<html></html>"}

	want := Exchange{
		URL: "http://app/callback?code=%3Credacted%3E&next=%2Fhome",
		RequestHeader: map[string]string{
			"Authorization": "<redacted>",
			"Cookie":        "<redacted>",
		},
		RequestBody: map[string]any{
			"name":     "Ana",
			"password": "<redacted>",
		},
		ResponseHeader: map[string]string{
			"Location": "/done?code=%3Credacted%3E",
		},
	}
	got := redact("http://app/callback?code=xyz&next=/home", request, response)

	assert.Equal(t, want, got)
}

func TestRedact_successWhenTheBodyIsRaw(t *testing.T) {
	request := Request{Body: []byte(`{"code":"xyz"}`)}
	response := _http.Response{}

	want := Exchange{URL: "http://app/token", RequestBody: map[string]any{"code": redacted}}
	got := redact("http://app/token", request, response)

	assert.Equal(t, want, got)
}

func TestRedactURL_success(t *testing.T) {
	want := "http://mock/authorize?STATE=%3Credacted%3E&client_id=neptune"
	got := redactURL("http://mock/authorize?client_id=neptune&STATE=abc")

	assert.Equal(t, want, got)
}

func TestRedactURL_successWhenTheValueIsNotAURL(t *testing.T) {
	want := redacted
	got := redactURL("http://app/%zz")

	assert.Equal(t, want, got)
}

func TestRedactHeader_success(t *testing.T) {
	header := map[string]string{"Set-Cookie": "session=abc", "Location": "/next?nonce=n", "Content-Type": "application/json"}

	want := map[string]string{"Set-Cookie": redacted, "Location": "/next?nonce=%3Credacted%3E", "Content-Type": "application/json"}
	got := redactHeader(header)

	assert.Equal(t, want, got)
}

func TestRedactHeader_successWhenThereIsNoHeader(t *testing.T) {
	got := redactHeader(map[string]string{})

	assert.Nil(t, got)
}

func TestRedactJSON_success(t *testing.T) {
	want := map[string]any{"sessions": []any{map[string]any{"id": "s1", "token": redacted}}}
	got := redactJSON([]byte(`{"sessions":[{"id":"s1","token":"abc"}]}`))

	assert.Equal(t, want, got)
}

func TestRedactJSON_successWhenTheBodyIsNotJSON(t *testing.T) {
	got := redactJSON([]byte("pong"))

	assert.Nil(t, got)
}

func TestRedactValue_success(t *testing.T) {
	value := []any{map[string]any{"ID_TOKEN": "abc", "data": map[string]any{"code": "xyz"}}, "plain"}

	want := []any{map[string]any{"ID_TOKEN": redacted, "data": map[string]any{"code": redacted}}, "plain"}
	got := redactValue(value)

	assert.Equal(t, want, got)
}
