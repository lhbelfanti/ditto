package log

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

// MockLogOutput redirects the package logger to the returned buffer at trace level and restores
// the default logger when the test ends.
func MockLogOutput(t *testing.T) *bytes.Buffer {
	t.Helper()
	buf := &bytes.Buffer{}
	NewCustomLogger(buf, zerolog.TraceLevel)
	t.Cleanup(func() { NewCustomLogger(os.Stdout, zerolog.DebugLevel) })
	return buf
}

// MockLogEntries returns the entries written to output, one per line, without their time.
func MockLogEntries(output *bytes.Buffer) []map[string]any {
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
