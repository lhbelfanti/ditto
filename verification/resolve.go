package verification

import (
	"net/url"
	"strings"
)

// resolve splits the route of request into its method and target, fills the placeholders of the target with its
// params and appends its query. A target that starts with "/" is a route of the app at baseURL.
func resolve(request Request, baseURL string) (string, string, error) {
	method, target, found := strings.Cut(request.Route, " ")
	if !found {
		return "", "", ErrRouteWithoutMethod
	}

	for name, value := range request.Params {
		placeholder := "{" + name + "}"
		if !strings.Contains(target, placeholder) {
			return "", "", ErrUnknownParam
		}

		target = strings.ReplaceAll(target, placeholder, url.PathEscape(value))
	}

	if strings.Contains(target, "{") {
		return "", "", ErrMissingParam
	}

	if strings.HasPrefix(target, "/") {
		target = strings.TrimSuffix(baseURL, "/") + target
	}

	if len(request.Query) > 0 {
		separator := "?"
		if strings.Contains(target, "?") {
			separator = "&"
		}

		target += separator + request.Query.Encode()
	}

	return method, target, nil
}
