package app

import "errors"

var (
	ErrCantOpenAdmin         = errors.New("app: can't open the administrator pool")
	ErrAdminUnavailable      = errors.New("app: the administrator pool is unreachable")
	ErrCantProvision         = errors.New("app: can't provision the service role and database")
	ErrInvalidDatabaseEnv    = errors.New("app: invalid database environment")
	ErrCantProvisionDatabase = errors.New("app: can't provision the database")
	ErrCantOpenDatabase      = errors.New("app: can't open the database")
	ErrCantReachDatabase     = errors.New("app: the database is unreachable")
	ErrCantApplyMigrations   = errors.New("app: can't apply the migrations")
	ErrInvalidPort           = errors.New("app: invalid port")
	ErrCantInitDatabase      = errors.New("app: can't initialize the database")
	ErrCantServe             = errors.New("app: can't serve")
)
