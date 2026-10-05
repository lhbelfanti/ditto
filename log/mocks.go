package log

import (
	"bytes"
	"os"
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
