package broker

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// declareQueue declares a durable queue on the given channel.
func declareQueue(ch *amqp091.Channel, name string) (amqp091.Queue, error) {
	q, err := ch.QueueDeclare(
		name,  // name
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return amqp091.Queue{}, fmt.Errorf("%w: %v", ErrFailedToDeclareQueue, err)
	}
	return q, nil
}
