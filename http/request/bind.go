package request

import (
	"encoding/json"
	"fmt"
	"maps"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"time"
)

// Bind reads the path params, the query params and the body of r, in that order, and returns the first error.
// A required param that is absent or equal to its zero value is missing; declare it Optional when the zero value is valid.
// A body that implements Validator is checked with IsValid, and its error is returned as it is. With NoBody the body is not read.
func Bind[B any](r *http.Request, path, query Fields) (Request[B], error) {
	pathValues, err := parse(path, r.PathValue, ErrMissingPathParam, ErrInvalidPathParam)
	if err != nil {
		return Request[B]{}, err
	}

	queryValues, err := parse(query, r.URL.Query().Get, ErrMissingQueryParam, ErrInvalidQueryParam)
	if err != nil {
		return Request[B]{}, err
	}

	request := Request[B]{Path: pathValues, Query: queryValues}

	_, noBody := any(request.Body).(NoBody)
	if noBody {
		return request, nil
	}

	err = json.NewDecoder(r.Body).Decode(&request.Body)
	if err != nil {
		return Request[B]{}, fmt.Errorf("%w: %w", ErrInvalidBody, err)
	}

	validator, validates := any(&request.Body).(Validator)
	if validates {
		err = validator.IsValid()
		if err != nil {
			return Request[B]{}, err
		}
	}

	return request, nil
}

// parse reads every declared param with lookup, in name order so that the first error is always the same one.
func parse(fields Fields, lookup func(string) string, errMissing, errInvalid error) (Values, error) {
	values := Values{fields: fields, values: make(map[string]any, len(fields))}

	for _, name := range slices.Sorted(maps.Keys(fields)) {
		field := fields[name]

		raw := strings.TrimSpace(lookup(name))
		if raw == "" {
			if field.optional {
				continue
			}

			return Values{}, fmt.Errorf("%w: %s", errMissing, name)
		}

		value, err := field.kind.parse(raw)
		if err != nil {
			return Values{}, fmt.Errorf("%w: %s: %w", errInvalid, name, err)
		}

		if !field.optional && reflect.ValueOf(value).IsZero() {
			return Values{}, fmt.Errorf("%w: %s", errMissing, name)
		}

		values.values[name] = value
	}

	return values, nil
}

// parse converts the raw text of a param into the type of the kind.
func (k kind) parse(raw string) (any, error) {
	switch k {
	case kindInt:
		return strconv.Atoi(raw)
	case kindBool:
		return strconv.ParseBool(raw)
	case kindTime:
		return time.Parse(time.RFC3339, raw)
	default:
		return raw, nil
	}
}
