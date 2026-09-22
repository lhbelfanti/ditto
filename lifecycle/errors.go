package lifecycle

import "errors"

var (
	ErrServeFailed          = errors.New("lifecycle: server failed unexpectedly")
	ErrShutdownFailed       = errors.New("lifecycle: http shutdown failed")
	ErrShutdownTimedOut     = errors.New("lifecycle: http shutdown timed out")
	ErrDatabaseCloseTimeout = errors.New("lifecycle: database close timed out")
)
