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
	assert.ErrorIs(t, err, underlying)
}

func TestRabbitMQBroker_CloseConnection_successWhenProcessingIsInFlight(t *testing.T) {
	messages := make(chan amqp091.Delivery, 1)
	acknowledger := MockAcknowledger(1, 1)
	messages <- amqp091.Delivery{DeliveryTag: 1, Acknowledger: acknowledger}
	b := &RabbitMQBroker{channel: &mockChannel{}, messages: messages}
	started := make(chan struct{})
	release := make(chan struct{})
	consumerDone := make(chan struct{})
	processor := MockBlockingProcessor(started, release, nil)
	go func() {
		b.InitMessageConsumerWithFunction(1, processor)
		close(consumerDone)
	}()
	select {
	case <-started:
	case <-time.After(2 * time.Second):
		t.Fatal("processor never started")
	}

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
