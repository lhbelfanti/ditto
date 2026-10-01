# broker

The `broker` package wraps `amqp091-go` behind a small `MessageBroker` interface for RabbitMQ
producer/consumer pairs.

## Concurrency and backpressure

`InitMessageConsumerWithFunction(concurrentMessages, processorFunc)` bounds processing two ways:
it sets the channel's QoS prefetch to `concurrentMessages` (so RabbitMQ itself slows deliveries),
and it bounds the number of concurrently running `processorFunc` goroutines to the same value via
an internal semaphore. Both exist together because QoS alone only limits how many *unacknowledged*
deliveries RabbitMQ will hand out — it does not, by itself, cap how many goroutines your own
process spawns to handle them.

The value must be greater than zero. Invalid values or a QoS setup failure are logged and stop
the consumer before it receives deliveries.

## Shutdown

`CloseConnection()` waits up to 10 seconds for any in-flight `processorFunc` goroutines spawned by
`InitMessageConsumerWithFunction` to finish, then closes the connection regardless — so a message
that's mid-processing when shutdown starts gets a real chance to `Ack`/`Nack` instead of being
silently abandoned. A producer-only broker (built by `NewProducer`) closes immediately, since it
never spawns any processing goroutines.
Shutdown also stops the consumer loop from dispatching new work before waiting for in-flight work.

## Message-loss semantics

Queues are declared `durable=true` with **no dead-letter exchange**. A `processorFunc` failure
calls `Nack(false, false)` — no requeue. In practice this means **a processing failure drops the
message permanently**, not "will be retried." If your use case needs retry-then-DLQ semantics,
build it into `processorFunc` itself (e.g. requeue to a separate retry queue on failure) — this
package does not provide it.

## Foot-gun: producer vs. consumer

`NewProducer` never sets the broker's `messages` channel. Calling
`InitMessageConsumerWithFunction` on a broker built by `NewProducer` used to range over that nil
channel and block forever, silently. It now detects this and logs
`ErrNotAConsumer` instead of hanging — but the right fix is still to build the broker with the
constructor matching how you intend to use it: `NewConsumer` to consume, `NewProducer` to publish.

## Testing

An internal channel interface lets unit tests exercise queue declaration, publishing, QoS,
concurrency limits, and Ack/Nack behavior without a running RabbitMQ server. Dial and constructor
success paths still need an integration test with RabbitMQ. `MockEnqueue` provides a
`MessageBroker` test double for consumer code that doesn't need to exercise this package's own
RabbitMQ wiring.
