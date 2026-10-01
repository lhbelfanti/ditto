package middleware

import "errors"

const ErrMsgInvalidToken string = "invalid or expired token"

var (
	ErrMissingAuthHeader = errors.New("middleware: authorization header required")
	ErrInvalidToken      = errors.New(ErrMsgInvalidToken)
)
