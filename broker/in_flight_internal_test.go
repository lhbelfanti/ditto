package broker

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWaitInFlight_success(t *testing.T) {
	var inFlight sync.WaitGroup
	done := make(chan struct{})

	waitInFlight(&inFlight, done)

	_, got := <-done

	assert.False(t, got)
}
