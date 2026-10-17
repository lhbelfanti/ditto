package request

import "errors"

var (
	ErrInvalidPath  = errors.New("invalid path")
	ErrInvalidQuery = errors.New("invalid query")
	ErrInvalidBody  = errors.New("invalid body")
	ErrMissingParam = errors.New("missing param")
	ErrInvalidParam = errors.New("invalid param")
	ErrNotAnInt     = errors.New("not an int")
	ErrNotABool     = errors.New("not a bool")
	ErrNotATime     = errors.New("not an RFC 3339 time")
	ErrInvalidDTO   = errors.New("invalid request DTO")
)
