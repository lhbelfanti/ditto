# migration

The `migration` package applies `*.sql` files from a directory against a `migrations` tracking
table, in lexicographic filename order, skipping files already applied — and gives a
`cmd/migrations`-style CLI everything it needs to report status and attribute failures without
leaking driver detail.

## Core types

| Type | Role |
|---|---|
| `Runner` | Applies every pending file under a lock. Built by `MakeRunner`. |
| `Status` | Reports every file as applied/pending, without mutating the database. Built by `MakeStatus`. |
| `Apply` | Wraps a `Runner` with credential-safe, file-attributed failure reporting. Built by `MakeApply`, which receives a `PendingFileFinder` (`MakePendingFileFinder`). |
| `Dispatch` | Runs the `apply` or `status` command named by its first argument. Built by `MakeDispatch`, which receives an `Apply` and a `StatusRunner` (`MakeStatusRunner`). |
| `Record` | One file's name and applied/pending state, as reported by `Status`. |

## Startup migrations

Services that declare a database run migrations at boot through `app.Run`, which is the normal
path: the process applies every pending file before it serves traffic, and exits with an error if
any file fails. Nothing has to be triggered by hand, and no migration endpoint is exposed over HTTP.

## Concurrency and atomicity

Each file is applied in its own transaction, together with its row in the `migrations` table:

- A transaction-scoped advisory lock (`pg_advisory_xact_lock`) serializes runners. Replicas that
  start together queue on it, and the first one to get it applies the file; the rest find it
  applied and skip it.
- A failing file rolls back completely, including any `CREATE TABLE` it ran before the failing
  statement, so a retry starts clean.

A migration file runs inside a transaction, so it cannot contain statements that Postgres rejects
there, such as `CREATE INDEX CONCURRENTLY` or `VACUUM`.

---

## End-to-end example: `cmd/migrations/main.go`

```go
pg, err := database.InitPostgres(ctx)
if err != nil {
    log.Fatalf("%s", err)
}
defer pg.Close()

runner := migration.MakeRunner(pg.Database(), "./migrations")
selectOneBool := database.MakeSelectOne[bool](pg.Database(), pgx.RowTo[bool])
selectNames := database.MakeSelect[string](pg.Database(), database.MakeCollectRows(pgx.RowTo[string]))
status := migration.MakeStatus(
    migration.MakeListFiles("./migrations"),
    migration.MakeTableExists(selectOneBool),
    migration.MakeAppliedNames(selectNames),
)
findPendingFile := migration.MakePendingFileFinder(status)
apply := migration.MakeApply(runner, findPendingFile)
runStatus := migration.MakeStatusRunner(status, os.Stdout)
dispatch := migration.MakeDispatch(apply, runStatus)

err = dispatch(ctx, os.Args[1:])
if err != nil {
    log.Fatalf("%s", err)
}
```

The CLI wires each operation explicitly in `main`: the database selects receive the connection, the
migration makers receive those selects, `MakeStatus` receives the three operations it uses, and
`MakeDispatch` receives the `apply` and `runStatus` it can run. No Maker builds another one inside its
logic.
The example needs an import of `github.com/jackc/pgx/v5` for `pgx.RowTo`.

### `MakeDispatch`'s two commands

- `apply` (also the default with no arguments) — applies every pending file exactly once.
  Re-running after everything is applied exits zero and applies nothing new.
- `status` — a read-only report, printed as `applied <filename>`/`pending <filename>` per file,
  safe to run before the first `apply` even against a genuinely clean database with no tracking
  table yet.

---

## Sentinel errors

| Error | Meaning |
|---|---|
| `ErrFailedToCreateTable` | The `migrations` tracking table couldn't be created. |
| `ErrFailedToExecute` | A migration file's SQL failed to execute. |
| `ErrUnableToReadFile` | The migrations directory couldn't be listed, or a matched file couldn't be read. |
| `ErrFailedToInsertApplied` / `ErrFailedToCheckApplied` | The tracking-table insert/select for one file failed. |
| `ErrFailedToCheckTableExists` / `ErrFailedToSelectAppliedNames` | `Status`'s own read-only checks failed. |
| `ErrFailedToBeginTransaction` / `ErrFailedToLockMigrations` | `MakeRunner` could not open a transaction or take the migration lock. |
| `ErrFailedToCommitTrackingTable` / `ErrFailedToCommitFile` | The commit of the migrations table creation, or of one file's transaction, failed. |
| `ErrFailedToApply` | `Apply`'s own sentinel — always present on an `*ApplyError`, alongside the original runner cause via `errors.Is`/`errors.As`. |
| `ErrUnknownCommand` | The `Dispatch` got anything other than no args, `apply`, or `status`. |

`*ApplyError` additionally carries, when available: the name of the first still-pending file
(`File`), an allowlisted PostgreSQL SQLSTATE (`Code`), and whether file attribution itself failed
(`AttributionUnavailable`) — never the raw underlying driver error, which may carry a
credential-bearing DSN.

```go
err := apply(ctx)
var applyErr *migration.ApplyError
if errors.As(err, &applyErr) && applyErr.File != "" {
    log.Errorf("migration %s failed (postgresql %s)", applyErr.File, applyErr.Code)
}
```
