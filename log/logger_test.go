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
	tests := []struct {
		name  string
		logFn func(ctx context.Context, msg string)
		level string
	}{
		{name: "Trace", logFn: log.Trace, level: "trace"},
		{name: "Debug", logFn: log.Debug, level: "debug"},
		{name: "Info", logFn: log.Info, level: "info"},
		{name: "Warn", logFn: log.Warn, level: "warn"},
		{name: "Error", logFn: log.Error, level: "error"},
		{name: "Fatal", logFn: log.Fatal, level: "fatal"},
		{name: "Panic", logFn: log.Panic, level: "panic"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := log.MockLogOutput(t)

			tt.logFn(context.Background(), "a message")

			want := map[string]any{"level": tt.level, "message": "a message"}
			entry := map[string]any{}
			_ = json.Unmarshal(buf.Bytes(), &entry)
			got := map[string]any{"level": entry["level"], "message": entry["message"]}

			assert.Equal(t, want, got)
		})
	}
}

func TestErr_success(t *testing.T) {
	buf := log.MockLogOutput(t)

	log.Err(context.Background(), assert.AnError, "error message")

	want := map[string]any{"level": "error", "message": "error message", "error": "assert.AnError general error for testing"}
	entry := map[string]any{}
	_ = json.Unmarshal(buf.Bytes(), &entry)
	got := map[string]any{"level": entry["level"], "message": entry["message"], "error": entry["error"]}

	assert.Equal(t, want, got)
}

func TestErr_successWhenErrorIsNil(t *testing.T) {
	buf := log.MockLogOutput(t)

	log.Err(context.Background(), nil, "info message")

	want := map[string]any{"level": "info", "message": "info message"}
	entry := map[string]any{}
	_ = json.Unmarshal(buf.Bytes(), &entry)
	got := map[string]any{"level": entry["level"], "message": entry["message"]}

	assert.Equal(t, want, got)
}
