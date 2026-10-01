package broker

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

// dial connects to RabbitMQ at the given URL, bounded by ctx. amqp091.Dial itself has no
// context-aware variant, so it runs in a goroutine raced against ctx.Done(); if ctx wins, any
// connection the dial still goes on to establish is closed in the background rather than leaked.
func dial(ctx context.Context, url string) (*amqp091.Connection, error) {
	type result struct {
		conn *amqp091.Connection
		err  error
	}

	done := make(chan result, 1)
	go func() {
		conn, err := amqp091.Dial(url)
		done <- result{conn, err}
	}()

	select {
	case r := <-done:
		if r.err != nil {
			return nil, fmt.Errorf("%w: %v", ErrFailedToConnect, r.err)
		}
		return r.conn, nil
	case <-ctx.Done():
		go func() {
			if r := <-done; r.conn != nil {
				r.conn.Close()
			}
		}()
		return nil, fmt.Errorf("%w: %v", ErrFailedToConnect, ctx.Err())
	}
}
