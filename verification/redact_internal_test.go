package verification

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedact_success(t *testing.T) {
	req, _ := http.NewRequest(http.MethodPost, "http://app/callback?code=xyz&next=/home", nil)
	req.Header.Set("Authorization", "Bearer session")
	req.Header.Set("Cookie", "state=abc")
	resp := &http.Response{Header: http.Header{"Location": {"/done?code=xyz"}}}

	want := Exchange{
		Method:         http.MethodPost,
		URL:            "http://app/callback?code=%3Credacted%3E&next=%2Fhome",
		RequestHeader:  map[string]string{"Authorization": redacted, "Cookie": redacted},
		RequestBody:    map[string]any{"name": "Ana", "password": redacted},
		ResponseHeader: map[string]string{"Location": "/done?code=%3Credacted%3E"},
	}
	got := redact(req, []byte(`{"name":"Ana","password":"pw"}`), resp, []byte("<html></html>"))

	assert.Equal(t, want, got)
}

func TestRedact_successWhenThereIsNoResponse(t *testing.T) {
	req, _ := http.NewRequest(http.MethodGet, "http://app/ping/v1", nil)

	want := Exchange{Method: http.MethodGet, URL: "http://app/ping/v1"}
	got := redact(req, nil, nil, nil)

	assert.Equal(t, want, got)
}

func TestFlatten_success(t *testing.T) {
	header := http.Header{"Vary": {"Origin", "Accept"}}

	want := map[string]string{"Vary": "Origin, Accept"}
	got := flatten(header)

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

func TestRedactValue_successWhenTheSensitiveValueIsNotAString(t *testing.T) {
	value := map[string]any{"code": float64(200), "data": map[string]any{"token": "abc"}}

	want := map[string]any{"code": float64(200), "data": map[string]any{"token": redacted}}
	got := redactValue(value)

	assert.Equal(t, want, got)
}
