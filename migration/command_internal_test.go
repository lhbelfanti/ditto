package migration

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStatusRunner_success(t *testing.T) {
	records := []Record{
		{Name: "000_foundation.sql", Applied: true},
		{Name: "001_second.sql", Applied: false},
	}
	mockStatus := MockStatus(records, nil)
	out := &bytes.Buffer{}

	runStatus := makeStatusRunner(mockStatus, out)

	want := "applied 000_foundation.sql\npending 001_second.sql\n"
	_ = runStatus(context.Background())
	got := out.String()

	assert.Equal(t, want, got)
}

func TestStatusRunner_failsWhenStatusFails(t *testing.T) {
	want := errors.New("status failed")
	mockStatus := MockStatus(nil, want)
	runStatus := makeStatusRunner(mockStatus, &bytes.Buffer{})

	got := runStatus(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestStatusRunner_failsWhenWriterFails(t *testing.T) {
	records := []Record{{Name: "000_foundation.sql", Applied: true}}
	mockStatus := MockStatus(records, nil)

	runStatus := makeStatusRunner(mockStatus, MockClosedWriter())

	want := io.ErrClosedPipe
	got := runStatus(context.Background())

	assert.ErrorIs(t, got, want)
}
