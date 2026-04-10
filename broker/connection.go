package broker

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// dial connects to RabbitMQ at the given URL.
func dial(url string) (*amqp091.Connection, error) {
	conn, err := amqp091.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToConnect, err)
	}
	return conn, nil
}
