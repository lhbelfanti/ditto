package migration

import "errors"

var (
	ErrFailedToCreateTable        = errors.New("migration: failed to create migrations table")
	ErrFailedToExecute            = errors.New("migration: failed to execute migration")
	ErrUnableToReadFile           = errors.New("migration: unable to read file")
	ErrFailedToInsertApplied      = errors.New("migration: failed to insert applied migration")
	ErrFailedToCheckApplied       = errors.New("migration: failed to check if migration is applied")
	ErrFailedToCheckTableExists   = errors.New("migration: failed to check if migrations table exists")
	ErrFailedToSelectAppliedNames = errors.New("migration: failed to select applied migration names")
	ErrFailedToApply              = errors.New("migration: failed to apply migrations")
)
