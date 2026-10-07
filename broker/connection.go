package broker

import (
	"context"
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

func closeLateDial(done <-chan dialResult) {
	r := <-done
	if r.conn != nil {
		r.conn.Close()
	}
}

// dial connects to RabbitMQ at the given URL, bounded by ctx. amqp091.Dial itself has no
// context-aware variant, so it runs in a goroutine raced against ctx.Done(); if ctx wins, any
// connection the dial still goes on to establish is closed in the background rather than leaked.
func dial(ctx context.Context, url string) (*amqp091.Connection, error) {
	done := make(chan dialResult, 1)
	go func() {
		conn, err := amqp091.Dial(url)
		done <- dialResult{conn, err}
	}()

	select {
	case r := <-done:
		if r.err != nil {
			return nil, fmt.Errorf("%w: %w", ErrFailedToConnect, r.err)
		}

		return r.conn, nil
	case <-ctx.Done():
		go closeLateDial(done)
		return nil, fmt.Errorf("%w: %w", ErrConnectCanceled, ctx.Err())
	}
}
