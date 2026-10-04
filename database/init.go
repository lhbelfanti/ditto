package database

import (
	"context"
	"os"
	"time"
)

// Init validates the POSTGRES_DB_* variables, opens the pool, and verifies connectivity within
// startupTimeout. The pool is the package-wide singleton from InitPostgres, so it is not closed on
// a failed check: a retry must still get a usable instance.
func Init(startupTimeout time.Duration) (*Postgres, error) {
	err := RequireEnv(os.LookupEnv)
	if err != nil {
		return nil, err
	}

	pg, err := InitPostgres()
	if err != nil {
		return nil, err
	}

	err = pg.MakeCheck(startupTimeout)(context.Background())
	if err != nil {
		return nil, err
	}

	return pg, nil
}
