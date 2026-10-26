package middleware

import "errors"

const (
	ErrMsgMissingAuthHeader string = "authorization header required"
	ErrMsgInvalidToken      string = "invalid or expired token"
	ErrMsgInvalidSession    string = "invalid or expired session"
	ErrMsgReadOnlySession   string = "read-only session cannot modify data"
)

var (
	ErrMissingAuthHeader = errors.New(ErrMsgMissingAuthHeader)
	ErrInvalidToken      = errors.New(ErrMsgInvalidToken)
	ErrInvalidSession    = errors.New(ErrMsgInvalidSession)
	ErrReadOnlySession   = errors.New(ErrMsgReadOnlySession)
)
