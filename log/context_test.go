package log_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/log"
)

func TestParam_success(t *testing.T) {
	want := "key"
	got := log.Param("key", "value").Key

	assert.Equal(t, want, got)
}

func TestParam_successWhenValueIsKept(t *testing.T) {
	want := "value"
	got := log.Param("key", "value").Value

	assert.Equal(t, want, got)
}
