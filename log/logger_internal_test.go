package log

import (
	"bytes"
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLogWithLevel_success(t *testing.T) {
	buf := &bytes.Buffer{}
	NewCustomLogger(buf, zerolog.DebugLevel)

	logWithLevel(context.Background(), zerolog.WarnLevel, "careful")

	want := []map[string]any{{"level": "warn", "message": "careful"}}
	got := MockLogEntries(buf)

	assert.Equal(t, want, got)
}

func TestLogWithLevel_successWhenContextCarriesParams(t *testing.T) {
	buf := &bytes.Buffer{}
	NewCustomLogger(buf, zerolog.DebugLevel)
	ctx := With(context.Background(), Param("request_id", "abc"))

	logWithLevel(ctx, zerolog.InfoLevel, "hello")

	want := []map[string]any{{"level": "info", "message": "hello", "request_id": "abc"}}
	got := MockLogEntries(buf)

	assert.Equal(t, want, got)
}

func TestLogWithLevel_successWhenLevelIsBelowThreshold(t *testing.T) {
	buf := &bytes.Buffer{}
	NewCustomLogger(buf, zerolog.ErrorLevel)

	logWithLevel(context.Background(), zerolog.InfoLevel, "filtered")

	want := ""
	got := buf.String()

	assert.Equal(t, want, got)
}
