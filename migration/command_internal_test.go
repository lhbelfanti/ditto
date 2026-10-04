package migration

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunStatus_success(t *testing.T) {
	records := []Record{
		{Name: "000_foundation.sql", Applied: true},
		{Name: "001_second.sql", Applied: false},
	}
	mockStatus := MockStatus(records, nil)
	out := &bytes.Buffer{}

	want := "applied 000_foundation.sql\npending 001_second.sql\n"
	_ = runStatus(context.Background(), mockStatus, out)
	got := out.String()

	assert.Equal(t, want, got)
}

func TestRunStatus_failsWhenStatusFails(t *testing.T) {
	want := errors.New("status failed")
	mockStatus := MockStatus(nil, want)

	got := runStatus(context.Background(), mockStatus, &bytes.Buffer{})

	assert.ErrorIs(t, got, want)
}

func TestRunStatus_failsWhenWriterFails(t *testing.T) {
	records := []Record{{Name: "000_foundation.sql", Applied: true}}
	mockStatus := MockStatus(records, nil)
	reader, writer := io.Pipe()
	_ = reader.Close()

	want := io.ErrClosedPipe
	got := runStatus(context.Background(), mockStatus, writer)

	assert.ErrorIs(t, got, want)
}
