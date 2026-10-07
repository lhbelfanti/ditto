package broker

import "errors"

// Sentinel errors for the broker package.
var (
	ErrFailedToConnect        = errors.New("failed to connect to rabbitmq")
	ErrConnectCanceled        = errors.New("rabbitmq connection attempt canceled before it completed")
	ErrFailedToOpenChannel    = errors.New("failed to open rabbitmq channel")
	ErrFailedToDeclareQueue   = errors.New("failed to declare rabbitmq queue")
	ErrFailedToConsumeQueue   = errors.New("failed to consume rabbitmq queue")
	ErrFailedToPublishMessage = errors.New("failed to publish message to rabbitmq")
	ErrFailedToSetQoS         = errors.New("failed to set rabbitmq channel qos")
	ErrInvalidConcurrency     = errors.New("concurrentMessages must be greater than zero")
	ErrNotAConsumer           = errors.New("broker: InitMessageConsumer called on a broker built by NewProducer")
	ErrConsumerChannelClosed  = errors.New("broker: delivery channel closed before CloseConnection was called")
)
