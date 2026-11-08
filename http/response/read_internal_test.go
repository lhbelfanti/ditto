package response

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNilWhenEmpty_success(t *testing.T) {
	header := http.Header{"Location": {"/next"}}

	want := http.Header{"Location": {"/next"}}
	got := nilWhenEmpty(header)

	assert.Equal(t, want, got)
}

func TestNilWhenEmpty_successWhenTheHeaderHasNoValues(t *testing.T) {
	got := nilWhenEmpty(http.Header{})

	assert.Nil(t, got)
}
