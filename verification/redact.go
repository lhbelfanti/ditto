package verification

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
)

const redacted string = "<redacted>"

var (
	sensitiveKeys = map[string]bool{
		"access_token":  true,
		"client_secret": true,
		"code":          true,
		"code_verifier": true,
		"id_token":      true,
		"nonce":         true,
		"password":      true,
		"refresh_token": true,
		"secret":        true,
		"state":         true,
		"token":         true,
	}
	sensitiveHeaders = map[string]bool{"authorization": true, "cookie": true, "set-cookie": true}
)

// redact builds the evidence of req, sent with requestBody, and of resp, read as responseBody, with the sensitive
// values replaced. resp is nil when the request got no answer. A body that is not JSON is left out, since a
// redirect page can repeat the values of its URL.
func redact(req *http.Request, requestBody []byte, resp *http.Response, responseBody []byte) Exchange {
	exchange := Exchange{
		Method:        req.Method,
		URL:           redactURL(req.URL.String()),
		RequestHeader: redactHeader(flatten(req.Header)),
		RequestBody:   redactJSON(requestBody),
	}
	if resp != nil {
		exchange.ResponseHeader = redactHeader(flatten(resp.Header))
		exchange.ResponseBody = redactJSON(responseBody)
	}

	return exchange
}

// flatten joins the values of each header with ", ".
func flatten(header http.Header) map[string]string {
	flat := make(map[string]string, len(header))
	for key, values := range header {
		flat[key] = strings.Join(values, ", ")
	}

	return flat
}

// redactURL replaces the values of the sensitive query params of raw, or all of raw when it is not a URL.
func redactURL(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return redacted
	}

	query := parsed.Query()
	for key := range query {
		if sensitiveKeys[strings.ToLower(key)] {
			query.Set(key, redacted)
		}
	}

	parsed.RawQuery = query.Encode()

	return parsed.String()
}

// redactHeader replaces the credentials and cookies of header, and the sensitive query params of its Location.
func redactHeader(header map[string]string) map[string]string {
	if len(header) == 0 {
		return nil
	}

	safe := make(map[string]string, len(header))
	for key, value := range header {
		name := strings.ToLower(key)
		switch {
		case sensitiveHeaders[name]:
			safe[key] = redacted
		case name == "location":
			safe[key] = redactURL(value)
		default:
			safe[key] = value
		}
	}

	return safe
}

// redactJSON decodes body and replaces its sensitive fields, or returns nil when body is not JSON.
func redactJSON(body []byte) any {
	var value any
	err := json.Unmarshal(body, &value)
	if err != nil {
		return nil
	}

	return redactValue(value)
}

// redactValue replaces the sensitive fields of value, at any depth.
func redactValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		for key, field := range typed {
			if sensitiveKeys[strings.ToLower(key)] {
				typed[key] = redacted
				continue
			}

			typed[key] = redactValue(field)
		}
	case []any:
		for i, item := range typed {
			typed[i] = redactValue(item)
		}
	}

	return value
}
