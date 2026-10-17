package request

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Bind reads r into a new D, a struct with up to three fields: Path, Query and Body.
// Path and Query are structs whose fields carry a `path:"name"` or `query:"name"` tag. A field is required:
// absent, or equal to its zero value, it is missing. A pointer field is optional, and stays nil when the param is absent.
// Body is decoded from JSON and, when it implements Validator, checked with IsValid, whose error is returned as it is.
// It reads path, query and body in that order, each in field order, and returns the first error.
func Bind[D any](r *http.Request) (D, error) {
	var dto D

	value := reflect.ValueOf(&dto).Elem()
	if value.Kind() != reflect.Struct {
		return dto, fmt.Errorf("%w: %s is not a struct", ErrInvalidDTO, value.Type())
	}

	path := value.FieldByName("Path")
	if path.IsValid() {
		err := parse(path, "path", r.PathValue, ErrMissingPathParam, ErrInvalidPathParam)
		if err != nil {
			return *new(D), err
		}
	}

	query := value.FieldByName("Query")
	if query.IsValid() {
		err := parse(query, "query", r.URL.Query().Get, ErrMissingQueryParam, ErrInvalidQueryParam)
		if err != nil {
			return *new(D), err
		}
	}

	body := value.FieldByName("Body")
	if !body.IsValid() {
		return dto, nil
	}

	err := json.NewDecoder(r.Body).Decode(body.Addr().Interface())
	if err != nil {
		return *new(D), fmt.Errorf("%w: %w", ErrInvalidBody, err)
	}

	validator, validates := body.Addr().Interface().(Validator)
	if validates {
		err = validator.IsValid()
		if err != nil {
			return *new(D), err
		}
	}

	return dto, nil
}

// parse fills every field of section with the param named by its tag, read with lookup.
func parse(section reflect.Value, tag string, lookup func(string) string, errMissing, errInvalid error) error {
	if section.Kind() != reflect.Struct {
		return fmt.Errorf("%w: %s is not a struct", ErrInvalidDTO, section.Type())
	}

	for i := range section.NumField() {
		field := section.Type().Field(i)

		name, tagged := field.Tag.Lookup(tag)
		if !tagged || !field.IsExported() {
			return fmt.Errorf("%w: %s needs to be exported and have a %s tag", ErrInvalidDTO, field.Name, tag)
		}

		target := section.Field(i)
		optional := target.Kind() == reflect.Pointer

		raw := strings.TrimSpace(lookup(name))
		if raw == "" {
			if optional {
				continue
			}

			return fmt.Errorf("%w: %s", errMissing, name)
		}

		if optional {
			target.Set(reflect.New(target.Type().Elem()))
			target = target.Elem()
		}

		err := set(target, raw)
		if err != nil {
			return fmt.Errorf("%w: %s: %w", errInvalid, name, err)
		}

		if !optional && target.IsZero() {
			return fmt.Errorf("%w: %s", errMissing, name)
		}
	}

	return nil
}

// set parses raw into target: a string (kept as it is), an int of any size, a bool or a time.Time (RFC 3339).
func set(target reflect.Value, raw string) error {
	if target.Type() == reflect.TypeFor[time.Time]() {
		parsed, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			return err
		}

		target.Set(reflect.ValueOf(parsed))

		return nil
	}

	switch target.Kind() {
	case reflect.String:
		target.SetString(raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		parsed, err := strconv.ParseInt(raw, 10, target.Type().Bits())
		if err != nil {
			return err
		}

		target.SetInt(parsed)
	case reflect.Bool:
		parsed, err := strconv.ParseBool(raw)
		if err != nil {
			return err
		}

		target.SetBool(parsed)
	default:
		return fmt.Errorf("%w: unsupported type %s", ErrInvalidDTO, target.Type())
	}

	return nil
}
