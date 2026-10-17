package broker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenChannel_success(t *testing.T) {
	mockConn := MockConnection(t)

	got, _ := openChannel(mockConn)

	assert.NotNil(t, got)
}

func TestOpenChannel_failsWhenConnectionIsClosed(t *testing.T) {
	mockConn := MockConnection(t)
	_ = mockConn.Close()

	want := ErrFailedToOpenChannel
	_, got := openChannel(mockConn)

	assert.ErrorIs(t, got, want)
}
