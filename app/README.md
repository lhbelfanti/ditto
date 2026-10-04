# app

The `app` package bootstraps a service: listen port, database, migrations at boot, system routes
and graceful shutdown. A service's `main.go` reduces to one call to `Run`.

## Usage

```go
package main

import (
	"net/http"

	"github.com/lhbelfanti/ditto/v2/app"
	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/lhbelfanti/ditto/v2/http/middleware"
	"github.com/lhbelfanti/ditto/v2/setup"
)

func main() {
	setup.Must(app.Run(app.Options{
		Name:          "nebula-earth",
		PortEnv:       "APP_INTERNAL_PORT",
		MigrationsDir: "./migrations",
		Routes: func(mux *http.ServeMux, pg *database.Postgres) {
			// the service's own endpoints
		},
		Wrap: func(next http.Handler) http.Handler {
			return middleware.RequestID(middleware.CORS()(next))
		},
	}))
}
```

## `Options`

| Field | Meaning |
|---|---|
| `Name` | Appears in the startup log line. |
| `PortEnv` | Variable holding the internal listen port, e.g. `APP_INTERNAL_PORT`. Required. |
| `MigrationsDir` | Enables the database. When set, `Run` requires the `POSTGRES_DB_*` variables, checks connectivity, applies pending migrations, and mounts `GET /database/ping/v1`. Leave it empty for a service without a database. |
| `Routes` | Registers the service's endpoints. `pg` is `nil` when the database is disabled. |
| `Wrap` | Optionally decorates the whole handler, e.g. with request ID or CORS. |

## What `Run` does, in order

1. Points the logger at stdout, at `info` level.
2. Reads the port from `PortEnv`.
3. If `MigrationsDir` is set: validates the database variables, opens the pool, checks
   connectivity within 5 seconds, and applies pending migrations.
4. Mounts `GET /ping/v1`, and `GET /database/ping/v1` when the database is enabled, then `Routes`.
5. Serves HTTP until `SIGINT` or `SIGTERM`, then shuts down gracefully within 5 seconds.

Every failure before step 5 is returned, not logged-and-exited, so the caller decides how to exit.
`setup.Must` in the example turns it into a non-zero exit.
