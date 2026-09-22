package database

import "errors"

var (
	ErrNoRows  = errors.New("database: no rows found")
	ErrQuery   = errors.New("database: query execution failed")
	ErrCollect = errors.New("database: failed to collect rows")

	ErrDatabaseUnavailable = errors.New("database: unavailable")
	ErrCantInitDatabase    = errors.New("database: can't initialize database")
)

// SafeError renders only its package-owned sentinel message, while still letting errors.Is
// identify both the sentinel and the original cause it hides — so driver errors that may carry
// a credential-bearing DSN never reach a log line or an HTTP response body.
type SafeError struct {
	sentinel error
	cause    error
}

func (e *SafeError) Error() string {
	return e.sentinel.Error()
}

func (e *SafeError) Unwrap() []error {
	return []error{e.sentinel, e.cause}
}

// WrapUnavailable hides cause behind the credential-safe ErrDatabaseUnavailable sentinel.
func WrapUnavailable(cause error) error {
	return &SafeError{sentinel: ErrDatabaseUnavailable, cause: cause}
}

// WrapInitFailure hides cause behind the credential-safe ErrCantInitDatabase sentinel.
func WrapInitFailure(cause error) error {
	return &SafeError{sentinel: ErrCantInitDatabase, cause: cause}
}
