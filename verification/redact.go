package verification

import (
	"encoding/json"
	"net/url"
	"strings"

	"github.com/lhbelfanti/ditto/v2/http"
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

// redact builds the evidence of a request sent to target and its response, with the sensitive values replaced.
// A body that is not JSON is left out, since a redirect page can repeat the values of its URL.
func redact(target string, request Request, response http.Response) Exchange {
	requestHeader := make(map[string]string, len(request.Header)+1)
	for key, value := range request.Header {
		requestHeader[key] = value
	}

	if request.Token != "" {
		requestHeader["Authorization"] = "Bearer " + request.Token
	}

	responseHeader := make(map[string]string, len(response.Header))
	for key, values := range response.Header {
		responseHeader[key] = strings.Join(values, ", ")
	}

	requestBody, raw := request.Body.([]byte)
	if !raw {
		requestBody, _ = json.Marshal(request.Body)
	}

	return Exchange{
		URL:            redactURL(target),
		RequestHeader:  redactHeader(requestHeader),
		RequestBody:    redactJSON(requestBody),
		ResponseHeader: redactHeader(responseHeader),
		ResponseBody:   redactJSON([]byte(response.Body)),
	}
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
