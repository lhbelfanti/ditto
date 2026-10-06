package database

import "errors"

var (
	ErrNoRows  = errors.New("database: no rows found")
	ErrQuery   = errors.New("database: query execution failed")
	ErrCollect = errors.New("database: failed to collect rows")

	ErrDatabaseUnavailable = errors.New("database: unavailable")
	ErrCantInitDatabase    = errors.New("database: can't initialize database")

	ErrAdminNotConfigured = errors.New("database: administrator credentials are not configured")
	ErrFailedToProvision  = errors.New("database: failed to provision the service database")
)

// WrapUnavailable hides cause behind the credential-safe ErrDatabaseUnavailable sentinel.
func WrapUnavailable(cause error) error {
	return &SafeError{sentinel: ErrDatabaseUnavailable, cause: cause}
}

// WrapInitFailure hides cause behind the credential-safe ErrCantInitDatabase sentinel.
func WrapInitFailure(cause error) error {
	return &SafeError{sentinel: ErrCantInitDatabase, cause: cause}
}

// WrapProvisionFailure hides cause behind the credential-safe ErrFailedToProvision sentinel.
func WrapProvisionFailure(cause error) error {
	return &SafeError{sentinel: ErrFailedToProvision, cause: cause}
}
