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

// New returns a Binder for r that reads nothing until WithPath, WithQuery or WithBody declare it.
func New(r *http.Request) Binder {
	return Binder{r: r}
}

// WithPath declares the path params the handler reads.
func (b Binder) WithPath(path Fields) Binder {
	b.path = path

	return b
}

// WithQuery declares the query params the handler reads.
func (b Binder) WithQuery(query Fields) Binder {
	b.query = query

	return b
}

// WithBody declares the pointer the JSON body is decoded into.
func (b Binder) WithBody(body any) Binder {
	b.body = body

	return b
}

// Bind reads the path params, the query params and the body, in that order, and returns the first error.
// A required param that is absent or equal to its zero value is missing; declare it Optional when the zero value is valid.
// A body that implements Validator is checked with IsValid, and its error is returned as it is.
func (b Binder) Bind() (Request, error) {
	path, err := parse(b.path, b.r.PathValue, ErrMissingPathParam, ErrInvalidPathParam, ErrUndeclaredPathParam)
	if err != nil {
		return Request{}, err
	}

	query, err := parse(b.query, b.r.URL.Query().Get, ErrMissingQueryParam, ErrInvalidQueryParam, ErrUndeclaredQueryParam)
	if err != nil {
		return Request{}, err
	}

	if b.body == nil {
		return Request{Path: path, Query: query}, nil
	}

	err = json.NewDecoder(b.r.Body).Decode(b.body)
	if err != nil {
		return Request{}, fmt.Errorf("%w: %w", ErrInvalidBody, err)
	}

	validator, validates := b.body.(Validator)
	if validates {
		err = validator.IsValid()
		if err != nil {
			return Request{}, err
		}
	}

	return Request{Path: path, Query: query}, nil
}

// parse reads every declared param with lookup, in name order so that the first error is always the same one.
func parse(fields Fields, lookup func(string) string, errMissing, errInvalid, errUndeclared error) (Values, error) {
	values := Values{fields: fields, values: make(map[string]any, len(fields)), errUndeclared: errUndeclared}

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
