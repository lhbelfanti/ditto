package migration

import "errors"

var (
	ErrFailedToCreateTable  = errors.New("migration: failed to create migrations table")
	ErrFailedToExecute      = errors.New("migration: failed to execute migration")
	ErrUnableToReadFile     = errors.New("migration: unable to read file")
	ErrFailedToInsertApplied = errors.New("migration: failed to insert applied migration")
	ErrFailedToCheckApplied = errors.New("migration: failed to check if migration is applied")
)
