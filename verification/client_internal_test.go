package verification

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStopAtRedirect_success(t *testing.T) {
	want := http.ErrUseLastResponse
	got := stopAtRedirect(nil, nil)

	assert.ErrorIs(t, got, want)
}
