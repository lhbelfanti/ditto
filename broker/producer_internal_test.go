package broker

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/assert"
)

func TestRabbitMQBroker_EnqueueMessage_success(t *testing.T) {
	ch := &mockChannel{}
	b := &RabbitMQBroker{channel: ch, queue: amqp091.Queue{Name: "work"}}

	err := b.EnqueueMessage(context.Background(), `{"job":1}`)

	assert.NoError(t, err)
	assert.Equal(t, []byte(`{"job":1}`), ch.published.Body)
	assert.Equal(t, "application/json", ch.published.ContentType)
}

func TestRabbitMQBroker_EnqueueMessage_failsWhenPublishFails(t *testing.T) {
	underlying := errors.New("unavailable")
	ch := &mockChannel{publishErr: underlying}
	b := &RabbitMQBroker{channel: ch, queue: amqp091.Queue{Name: "work"}}

	err := b.EnqueueMessage(context.Background(), "{}")

	assert.ErrorIs(t, err, ErrFailedToPublishMessage)
}

func TestRabbitMQBroker_CloseConnection_successWhenProcessingIsInFlight(t *testing.T) {
	messages := make(chan amqp091.Delivery, 1)
	acknowledger := &mockAcknowledger{acks: make(chan uint64, 1), nacks: make(chan uint64, 1)}
	messages <- amqp091.Delivery{DeliveryTag: 1, Acknowledger: acknowledger}
	b := &RabbitMQBroker{channel: &mockChannel{}, messages: messages}
	started := make(chan struct{})
	release := make(chan struct{})
	consumerDone := make(chan struct{})
	go func() {
		b.InitMessageConsumerWithFunction(1, func(context.Context, []byte) error {
			close(started)
			<-release
			return nil
		})
		close(consumerDone)
	}()
	<-started

	closed := make(chan struct{})
	go func() {
		b.CloseConnection()
		close(closed)
	}()
	select {
	case <-closed:
		t.Fatal("CloseConnection returned before in-flight processing finished")
	case <-time.After(50 * time.Millisecond):
	}
	close(release)
	select {
	case <-closed:
	case <-time.After(2 * time.Second):
		t.Fatal("CloseConnection did not finish after processing completed")
	}
	select {
	case <-consumerDone:
	case <-time.After(2 * time.Second):
		t.Fatal("consumer remained blocked after shutdown")
	}

	assert.Equal(t, uint64(1), <-acknowledger.acks)
}
