package request

import "errors"

var (
	ErrMissingPathParam  = errors.New("missing path param")
	ErrInvalidPathParam  = errors.New("invalid path param")
	ErrMissingQueryParam = errors.New("missing query param")
	ErrInvalidQueryParam = errors.New("invalid query param")
	ErrInvalidBody       = errors.New("invalid request body")
)
