package log

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type (
	// field represent a key-value tuple that will be added to the context
	field struct {
		Key   string
		Value any
		embed bool
	}

	// FieldsError is an error that carries the log fields of the context it was created with. Its
	// text is the cause's text followed by the fields, so the fields travel up the call chain
	// with the error and show up wherever it is finally logged.
	FieldsError struct {
		cause  error
		fields map[string]any
	}

	logCtxKey struct{}

	// embedCtxKey keys the set of context fields that NewFieldsError embeds into an error
	embedCtxKey struct{}
)

// Error renders the cause followed by the fields in alphabetical order, e.g. `can't fetch [symbol=BTC/USDT timeframe=1h]`.
func (e *FieldsError) Error() string {
	keys := make([]string, 0, len(e.fields))
	for key := range e.fields {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	parts := make([]string, len(keys))
	for i, key := range keys {
		parts[i] = key + "=" + formatField(e.fields[key])
	}

	return e.cause.Error() + " [" + strings.Join(parts, " ") + "]"
}

// Unwrap returns the cause, so errors.Is and errors.As keep working through a FieldsError.
func (e *FieldsError) Unwrap() error {
	return e.cause
}

// formatField renders a field value for an error text. Times use RFC 3339.
func formatField(value any) string {
	if t, ok := value.(time.Time); ok {
		return t.Format(time.RFC3339)
	}

	return fmt.Sprint(value)
}
