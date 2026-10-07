package database

import "errors"

var (
	ErrNoRows  = errors.New("database: no rows found")
	ErrQuery   = errors.New("database: query execution failed")
	ErrCollect = errors.New("database: failed to collect rows")

	ErrDatabaseUnavailable = errors.New("database: unavailable")
	ErrCantInitDatabase    = errors.New("database: can't initialize database")

	ErrAdminNotConfigured = errors.New("database: administrator credentials are not configured")

	ErrFailedToCreateRole       = errors.New("database: failed to create the service role")
	ErrFailedToSyncPassword     = errors.New("database: failed to sync the service role password")
	ErrFailedToCreateDatabase   = errors.New("database: failed to create the service database")
	ErrFailedToRevokePublic     = errors.New("database: failed to revoke public access to the service database")
	ErrFailedToBuildStatement   = errors.New("database: failed to build the provisioning statement")
	ErrFailedToExecuteStatement = errors.New("database: failed to execute the provisioning statement")
)

// WrapUnavailable hides cause behind the credential-safe ErrDatabaseUnavailable sentinel.
func WrapUnavailable(cause error) error {
	return &SafeError{sentinel: ErrDatabaseUnavailable, cause: cause}
}

// WrapInitFailure hides cause behind the credential-safe ErrCantInitDatabase sentinel.
func WrapInitFailure(cause error) error {
	return &SafeError{sentinel: ErrCantInitDatabase, cause: cause}
}
