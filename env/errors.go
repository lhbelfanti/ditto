package env

import "errors"

const (
	ErrMsgMissingKey  string = "env: missing required environment variable"
	ErrMsgInvalidPort string = "env: must be a port between 1 and 65535"
)

var (
	ErrMissingKey  = errors.New(ErrMsgMissingKey)
	ErrInvalidPort = errors.New(ErrMsgInvalidPort)
)
