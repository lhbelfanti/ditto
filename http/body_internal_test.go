package http

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/log"
)

func TestCloseBody_success(t *testing.T) {
	output, restore := log.Capture()
	t.Cleanup(restore)
	mockBody := io.NopCloser(strings.NewReader("done"))

	closeBody(context.Background(), mockBody)

	got := output.Entries()

	assert.Empty(t, got)
}

func TestCloseBody_successWhenCloseFails(t *testing.T) {
	output, restore := log.Capture()
	t.Cleanup(restore)
	mockBody := MockFailingBody{Reader: strings.NewReader("done"), Err: errors.New("boom")}

	closeBody(context.Background(), mockBody)

	want := []map[string]any{{"level": "warn", "message": "failed to close the response body", "cause": "boom"}}
	got := output.Entries()

	assert.Equal(t, want, got)
}
