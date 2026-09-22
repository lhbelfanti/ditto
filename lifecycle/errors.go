package lifecycle

import "errors"

const (
	ErrMsgServeFailed          string = "lifecycle: server failed unexpectedly"
	ErrMsgShutdownFailed       string = "lifecycle: http shutdown failed"
	ErrMsgShutdownTimedOut     string = "lifecycle: http shutdown timed out"
	ErrMsgDatabaseCloseTimeout string = "lifecycle: database close timed out"
)

var (
	ErrServeFailed          = errors.New(ErrMsgServeFailed)
	ErrShutdownFailed       = errors.New(ErrMsgShutdownFailed)
	ErrShutdownTimedOut     = errors.New(ErrMsgShutdownTimedOut)
	ErrDatabaseCloseTimeout = errors.New(ErrMsgDatabaseCloseTimeout)
)
