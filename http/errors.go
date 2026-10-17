package http

import "errors"

const ErrMsgDatabaseUnavailable string = "database unreachable"

var (
	ErrFailedToMarshalBody    = errors.New("failed to marshal body")
	ErrFailedToCreateRequest  = errors.New("failed to create request")
	ErrFailedToExecuteRequest = errors.New("failed to execute request")
	ErrFailedToReadResponse   = errors.New("failed to read response")
	ErrFailedToServe          = errors.New("failed to serve")
	ErrFailedToShutdown       = errors.New("failed to shut down gracefully")
	ErrCantListen             = errors.New("can't listen")

	ErrDatabaseUnavailable = errors.New(ErrMsgDatabaseUnavailable)
)

var (
	// Deprecated: use ErrFailedToMarshalBody.
	FailedToMarshalBody = ErrFailedToMarshalBody

	// Deprecated: use ErrFailedToCreateRequest.
	FailedToCreateRequest = ErrFailedToCreateRequest

	// Deprecated: use ErrFailedToExecuteRequest.
	FailedToExecuteRequest = ErrFailedToExecuteRequest

	// Deprecated: use ErrFailedToReadResponse.
	FailedToReadResponse = ErrFailedToReadResponse
)
