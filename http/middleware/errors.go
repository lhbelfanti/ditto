package middleware

import "errors"

const (
	ErrMsgMissingAuthHeader string = "authorization header required"
	ErrMsgInvalidToken      string = "invalid or expired token"
)

var (
	ErrMissingAuthHeader = errors.New(ErrMsgMissingAuthHeader)
	ErrInvalidToken      = errors.New(ErrMsgInvalidToken)
)
