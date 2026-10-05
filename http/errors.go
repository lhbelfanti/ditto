package http

import "errors"

const ErrMsgDatabaseUnavailable string = "database unreachable"

var (
	FailedToMarshalBody    = errors.New("failed to marshal body")
	FailedToCreateRequest  = errors.New("failed to create request")
	FailedToExecuteRequest = errors.New("failed to execute request")
	FailedToReadResponse   = errors.New("failed to read response")

	ErrDatabaseUnavailable = errors.New(ErrMsgDatabaseUnavailable)
)
