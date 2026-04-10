package broker

import "errors"

// Sentinel errors for the broker package.
var (
	ErrFailedToConnect        = errors.New("failed to connect to rabbitmq")
	ErrFailedToOpenChannel    = errors.New("failed to open rabbitmq channel")
	ErrFailedToDeclareQueue   = errors.New("failed to declare rabbitmq queue")
	ErrFailedToConsumeQueue   = errors.New("failed to consume rabbitmq queue")
	ErrFailedToPublishMessage = errors.New("failed to publish message to rabbitmq")
)
