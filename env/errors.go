package env

import "errors"

var (
	ErrMissingKey  = errors.New("env: missing required environment variable")
	ErrInvalidPort = errors.New("env: must be a port between 1 and 65535")
)
