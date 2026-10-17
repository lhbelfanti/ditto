package request

import "errors"

var (
	ErrMissingPathParam     = errors.New("missing path param")
	ErrInvalidPathParam     = errors.New("invalid path param")
	ErrUndeclaredPathParam  = errors.New("undeclared path param")
	ErrMissingQueryParam    = errors.New("missing query param")
	ErrInvalidQueryParam    = errors.New("invalid query param")
	ErrUndeclaredQueryParam = errors.New("undeclared query param")
	ErrInvalidBody          = errors.New("invalid request body")
)
