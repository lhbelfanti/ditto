# migration

The `migration` package applies `*.sql` files from a directory against a `migrations` tracking
table, in lexicographic filename order, skipping files already applied — and gives a
`cmd/migrations`-style CLI everything it needs to report status and attribute failures without
leaking driver detail.

## Core types

| Type | Role |
|---|---|
| `Runner` | Applies every pending file. Built by `MakeRunner`/`MakeRunnerWithDeps`. |
| `Status` | Reports every file as applied/pending, without mutating the database. Built by `MakeStatus`. |
| `Apply` | Wraps a `Runner` with credential-safe, file-attributed failure reporting. Built by `MakeApply`. |
| `Record` | One file's name and applied/pending state, as reported by `Status`. |

`Runner`'s underlying type (`func(ctx context.Context) error`) matches `http.MigrationRunner` —
assignable with an explicit conversion when registering the HTTP migrations route, with no import
of `http` needed by this package.

---

## End-to-end example: `cmd/migrations/main.go`

```go
pg, err := database.InitPostgres()
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
apply := migration.MakeApply(runner, status)

err = migration.Dispatch(ctx, os.Args[1:], apply, status, os.Stdout)
if err != nil {
    log.Fatalf("%s", err)
}
```

The CLI wires each operation explicitly: the database selects receive the connection, the
migration makers receive those selects, and `MakeStatus` receives the three operations it uses.
The example needs an import of `github.com/jackc/pgx/v5` for `pgx.RowTo`.

### Wiring the HTTP migrations route

```go
mux := http.NewServeMux()
dittohttp.RegisterSystemRoutes(mux).WithMigrationRunner(dittohttp.MigrationRunner(runner))
```

### `Dispatch`'s two commands

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
| `ErrFailedToApply` | `Apply`'s own sentinel — always present on an `*ApplyError`, alongside the original runner cause via `errors.Is`/`errors.As`. |
| `ErrUnknownCommand` | `Dispatch` got anything other than no args, `apply`, or `status`. |

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
