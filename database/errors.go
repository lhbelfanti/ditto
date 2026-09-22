package database

import "errors"

const (
	ErrMsgDatabaseUnavailable string = "database: unavailable"
	ErrMsgCantInitDatabase    string = "database: can't initialize database"
)

var (
	// ErrNoRows is returned when a query finds no matching rows.
	ErrNoRows = errors.New("database: no rows found")

	// ErrQuery is returned when a query execution fails.
	ErrQuery = errors.New("database: query execution failed")

	// ErrCollect is returned when row collection fails after a successful query.
	ErrCollect = errors.New("database: failed to collect rows")

	// ErrDatabaseUnavailable is returned when a bounded Check fails to reach the database.
	ErrDatabaseUnavailable = errors.New(ErrMsgDatabaseUnavailable)

	// ErrCantInitDatabase is returned when InitPostgres fails to establish a connection pool.
	ErrCantInitDatabase = errors.New(ErrMsgCantInitDatabase)
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
