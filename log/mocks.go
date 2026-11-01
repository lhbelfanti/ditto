package log

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/rs/zerolog"
)

// MockLogCapture holds what the package logger writes while a test runs.
type MockLogCapture struct {
	bytes.Buffer
}

// MockLogOutput redirects the package logger to the returned capture at trace level and restores
// the default logger when the test ends.
func MockLogOutput(t *testing.T) *MockLogCapture {
	t.Helper()
	capture := &MockLogCapture{}
	NewCustomLogger(capture, zerolog.TraceLevel)
	t.Cleanup(func() { NewCustomLogger(os.Stdout, zerolog.DebugLevel) })
	return capture
}

// Entries returns the entries the logger wrote, one per line, without their time.
func (capture *MockLogCapture) Entries() []map[string]any {
	lines := strings.Split(strings.TrimSpace(capture.String()), "\n")
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
