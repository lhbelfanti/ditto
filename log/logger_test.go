package log_test

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/log"
)

func TestNewCustomLogger_successWhenWriterIsNil(t *testing.T) {
	var buf bytes.Buffer
	log.NewCustomLogger(&buf, zerolog.TraceLevel)
	log.NewCustomLogger(nil, zerolog.TraceLevel)

	log.Info(context.Background(), "test message")

	want := ""
	got := buf.String()

	assert.Equal(t, want, got)
}

func TestLogLevels_success(t *testing.T) {
	type levelLogger func(ctx context.Context, msg string)

	tests := []struct {
		name   string
		logger levelLogger
		level  string
	}{
		{name: "Trace", logger: log.Trace, level: "trace"},
		{name: "Debug", logger: log.Debug, level: "debug"},
		{name: "Info", logger: log.Info, level: "info"},
		{name: "Warn", logger: log.Warn, level: "warn"},
		{name: "Error", logger: log.Error, level: "error"},
		{name: "Fatal", logger: log.Fatal, level: "fatal"},
		{name: "Panic", logger: log.Panic, level: "panic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := log.MockLogOutput(t)

			tt.logger(context.Background(), "a message")

			entry := map[string]any{}
			_ = json.Unmarshal(buf.Bytes(), &entry)

			want := map[string]any{"level": tt.level, "message": "a message"}
			got := map[string]any{"level": entry["level"], "message": entry["message"]}

			assert.Equal(t, want, got)
		})
	}
}

func TestErr_success(t *testing.T) {
	buf := log.MockLogOutput(t)

	log.Err(context.Background(), assert.AnError, "error message")

	entry := map[string]any{}
	_ = json.Unmarshal(buf.Bytes(), &entry)

	want := map[string]any{"level": "error", "message": "error message", "error": "assert.AnError general error for testing"}
	got := map[string]any{"level": entry["level"], "message": entry["message"], "error": entry["error"]}

	assert.Equal(t, want, got)
}

func TestErr_successWhenErrorIsNil(t *testing.T) {
	buf := log.MockLogOutput(t)

	log.Err(context.Background(), nil, "info message")

	entry := map[string]any{}
	_ = json.Unmarshal(buf.Bytes(), &entry)

	want := map[string]any{"level": "info", "message": "info message"}
	got := map[string]any{"level": entry["level"], "message": entry["message"]}

	assert.Equal(t, want, got)
}
