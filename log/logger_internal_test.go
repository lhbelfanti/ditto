package log

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestLogWithLevel_success(t *testing.T) {
	buf := &bytes.Buffer{}
	NewCustomLogger(buf, zerolog.DebugLevel)

	logWithLevel(context.Background(), zerolog.WarnLevel, "careful")

	want := map[string]any{"level": "warn", "message": "careful"}
	got := map[string]any{}
	_ = json.Unmarshal(buf.Bytes(), &got)
	delete(got, "time")

	assert.Equal(t, want, got)
}

func TestLogWithLevel_successWhenContextCarriesParams(t *testing.T) {
	buf := &bytes.Buffer{}
	NewCustomLogger(buf, zerolog.DebugLevel)
	ctx := With(context.Background(), Param("request_id", "abc"))

	logWithLevel(ctx, zerolog.InfoLevel, "hello")

	want := "abc"
	got := map[string]any{}
	_ = json.Unmarshal(buf.Bytes(), &got)

	assert.Equal(t, want, got["request_id"])
}

func TestLogWithLevel_successWhenLevelIsBelowThreshold(t *testing.T) {
	buf := &bytes.Buffer{}
	NewCustomLogger(buf, zerolog.ErrorLevel)

	logWithLevel(context.Background(), zerolog.InfoLevel, "filtered")

	want := ""
	got := buf.String()

	assert.Equal(t, want, got)
}
