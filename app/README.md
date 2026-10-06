# app

The `app` package bootstraps a service: listen port, database, migrations at boot, system routes
and graceful shutdown. A service's `main.go` reduces to one call to `Run`.

## Usage

```go
package main

import (
	"context"
	"net/http"

	"github.com/lhbelfanti/ditto/v2/app"
	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/lhbelfanti/ditto/v2/http/middleware"
	"github.com/lhbelfanti/ditto/v2/setup"
)

func main() {
	/* --- Setup --- */
	ctx := context.Background()

	pg := setup.Init(app.InitDatabase(ctx, "./migrations"))
	defer pg.Close()
	db := pg.Database()

	/* --- Dependencies by endpoint --- */
	// GET /candles/v1
	selectCandlesOp := database.MakeSelect[candle.DAO](db, collectCandles)
	getCandles := candle.MakeGet(selectCandlesOp)

	/* --- Routes --- */
	mux := http.NewServeMux()
	mux.HandleFunc("GET /candles/v1", getCandles)

	/* --- Server --- */
	options := app.Options{Name: "Nebula Saturn", Mux: mux, Database: pg}.
		WithMiddleware(middleware.RequestID).
		WithMiddleware(middleware.CORS())

	setup.Must(app.Run(options))
}
```

The `main` wires every dependency top-down; `Run` only serves the result. A service without a
database skips `InitDatabase` and leaves `Database` unset.

## `InitDatabase`

`InitDatabase(ctx, migrationsDir)` validates the `POSTGRES_DB_*` variables, opens the pool, checks
connectivity within 5 seconds and applies pending migrations. It returns the `*database.Postgres`
pool: `pg.Database()` is the `database.Connection` to inject into the makers, and the caller closes
the pool. Any failure is returned, and the pool is closed before returning.

## `Options`

| Field / method | Meaning |
|---|---|
| `Name` | Appears in the startup log line, e.g. `Nebula Saturn`. |
| `PortEnv` | Variable holding the internal listen port. Defaults to `APP_INTERNAL_PORT`. |
| `Mux` | The service's endpoints. `Run` mounts its system routes on it. |
| `Database` | The pool from `InitDatabase`. When set, `Run` mounts `GET /database/ping/v1`. |
| `Timeouts` | Bounds the ping route and shutdown. Zero fields use their defaults. |
| `WithMiddleware(m)` | Adds a middleware around the whole handler. The first one added is the outermost. |

`WithMiddleware` returns a copy, so an `Options` value can be shared and extended safely. The
`MigrationsDir`, `Routes` and `Wrap` fields still work but are deprecated: they make `Run` open the
database itself, which keeps the dependencies out of `main`.

## What `Run` does, in order

1. Points the logger at stdout, at `info` level.
2. Reads the port from `PortEnv`.
3. Mounts `GET /ping/v1` on `Mux`, and `GET /database/ping/v1` when `Database` is set.
4. Serves HTTP until `SIGINT` or `SIGTERM`, then shuts down gracefully within 5 seconds.

Every failure before step 4 is returned, not logged-and-exited, so the caller decides how to exit.
`setup.Must` in the example turns it into a non-zero exit.
