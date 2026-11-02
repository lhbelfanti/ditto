package log

import (
	"bytes"
	"encoding/json"
	"sort"
	"strings"
)

type (
	// MemoryOutput is a logger output that keeps what the logger writes in memory.
	MemoryOutput struct {
		bytes.Buffer
	}

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

	// logCtxKey keys the log fields stored in a context.
	logCtxKey struct{}

	// embedCtxKey keys the set of context fields that NewFieldsError embeds into an error
	embedCtxKey struct{}
)

// Entries returns the entries the logger wrote, one per line, without their time.
func (output *MemoryOutput) Entries() []map[string]any {
	lines := strings.Split(strings.TrimSpace(output.String()), "\n")
	entries := make([]map[string]any, 0, len(lines))
	for _, line := range lines {
		entry := map[string]any{}
		err := json.Unmarshal([]byte(line), &entry)
		if err != nil {
			continue
		}

		delete(entry, "time")
		entries = append(entries, entry)
	}

	return entries
}

// Error renders the cause followed by the fields in alphabetical order, e.g. `can't fetch [symbol=BTC/USDT timeframe=1h]`.
func (e *FieldsError) Error() string {
	keys := make([]string, 0, len(e.fields))
	for key := range e.fields {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, key+"="+formatField(e.fields[key]))
	}

	return e.cause.Error() + " [" + strings.Join(parts, " ") + "]"
}

// Unwrap returns the cause, so errors.Is and errors.As keep working through a FieldsError.
func (e *FieldsError) Unwrap() error {
	return e.cause
}
