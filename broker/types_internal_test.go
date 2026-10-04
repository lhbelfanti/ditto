package broker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRabbitMQBroker_stopChannel_success(t *testing.T) {
	b := &RabbitMQBroker{}

	first := b.stopChannel()
	second := b.stopChannel()

	want := first
	got := second

	assert.Equal(t, want, got)
}
