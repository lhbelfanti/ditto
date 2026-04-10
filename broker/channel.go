package broker

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// openChannel opens a channel on the given connection.
func openChannel(conn *amqp091.Connection) (*amqp091.Channel, error) {
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToOpenChannel, err)
	}
	return ch, nil
}
