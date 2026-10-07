# database

The `database` package provides a thin abstraction over `pgx/v5` for PostgreSQL. It exposes named generic function types for each SQL operation, so that a domain function's dependency signature immediately tells you which kind of query it runs.

## Core types

| Type | SQL operation | pgx primitive used |
|---|---|---|
| `Select[T]` | Multi-row `SELECT` | `db.Query` + `CollectRows[T]` |
| `SelectOne[T]` | Single-row `SELECT` | `db.Query` + `pgx.CollectOneRow` |
| `Insert[T]` | `INSERT … RETURNING` (scalar result) | `db.QueryRow` + `Scan` |
| `Delete` | `DELETE` | `db.Exec` |
| `Update` | `UPDATE` (also no-return `INSERT`) | `db.Exec` |
| `ExecFormatted` | Runs the SQL a query builds with `format()` (DDL with quoted identifiers) | `SelectOne[string]` + `Update` |
| `CollectRows[T]` | Row scanner passed to `Select[T]` | `pgx.CollectRows` |

## Connection settings

`InitPostgres(ctx)` reads these variables and opens a new pool on every call, so call it once per process. `RequireEnv` validates them before any connection attempt.

| Variable | Required | Notes |
|---|---|---|
| `POSTGRES_DB_HOST` | No | Defaults to `postgres_db`, the compose service name. Set it to point at a shared instance. |
| `POSTGRES_DB_PORT` | Yes | Valid TCP port. |
| `POSTGRES_DB_NAME`, `POSTGRES_DB_USER`, `POSTGRES_DB_PASS` | Yes | Non-empty. |

The credentials are escaped when the connection URL is built, so a password may hold spaces,
`@`, `/`, `:` or `=`.

### Creating the database

`InitPostgres` never creates anything: the database and its role must exist. To let a service
create its own, set the administrator credentials and call `MakeProvision`:

| Variable | Required | Notes |
|---|---|---|
| `POSTGRES_ADMIN_USER`, `POSTGRES_ADMIN_PASS` | No | A role allowed to create roles and databases. `OpenAdmin` connects with it to the `postgres` maintenance database, on the same host and port. It returns `ErrAdminNotConfigured` when either is empty. |

```go
admin, err := database.OpenAdmin(ctx)
if err != nil {
    return err
}
defer admin.Close()

selectStatement := database.MakeSelectOne[string](admin.Database(), pgx.RowTo[string])
update := database.MakeUpdate(admin.Database())
execFormatted := database.MakeExecFormatted(selectStatement, update)

provision := database.MakeProvision(execFormatted, database.TargetFromEnv())
err = provision(ctx)
```

`MakeProvision` creates the role and the database when they are missing, makes the role the
owner, syncs the role's password, and revokes the database from `PUBLIC` so no other role can
connect to it. It is idempotent. Each step fails with its own sentinel, wrapping the cause, so the
error says which step failed (see below). `app.InitDatabase` does all of this when the admin
variables are set.

---

## End-to-end example

### 1. Initialize the connection

```go
pg, err := database.InitPostgres(ctx)
if err != nil {
    log.Fatal(ctx, err.Error())
}
defer pg.Close()

db := pg.Database() // *pgxpool.Pool — satisfies database.Connection
```

### 2. Create generic operations in `main.go`

```go
// CollectRows scanner for corpus.DAO (fields mapped by position)
collectCorpusRows := database.MakeCollectRows[corpus.DAO](nil)

// Generic operations — created once, injected everywhere
selectCorpus   := database.MakeSelect[corpus.DAO](db, collectCorpusRows)
selectOneUser  := database.MakeSelectOne[user.DAO](db, nil)
insertCorpus   := database.MakeInsert[int](db)
deleteCorpus   := database.MakeDelete(db)
updateExec     := database.MakeUpdate(db)
```

> **Tip:** `MakeSelectOne` accepts a custom `pgx.RowToFunc[T]` as the second argument.
> Pass `nil` to use the default `pgx.RowToStructByPos[T]`, or provide your own function
> to scan scalar types (`bool`, `int`, …) or structs with non-positional mapping.

### 3. Domain factory functions

Each domain function accepts exactly one generic operation as its DB dependency.
The dependency type tells the reader which SQL operation is performed.

```go
// corpus/select.go
type SelectAll func(ctx context.Context) ([]DAO, error)

func MakeSelectAll(sel database.Select[DAO]) SelectAll {
    const query = `SELECT id, tweet_author, ... FROM corpus`
    return func(ctx context.Context) ([]DAO, error) {
        return sel(ctx, query)
    }
}

// corpus/insert.go
type Insert func(ctx context.Context, entry DTO) (int, error)

func MakeInsert(ins database.Insert[int]) Insert {
    const query = `INSERT INTO corpus(...) VALUES (...) RETURNING id`
    return func(ctx context.Context, entry DTO) (int, error) {
        return ins(ctx, query,
            entry.TweetAuthor, entry.TweetText, entry.Categorization,
        )
    }
}

// corpus/delete.go
type DeleteAll func(ctx context.Context) error

func MakeDeleteAll(del database.Delete) DeleteAll {
    const query = `DELETE FROM corpus`
    return func(ctx context.Context) error {
        return del(ctx, query)
    }
}
```

### 4. Wire in `main.go`

```go
corpusSelectAll := corpus.MakeSelectAll(selectCorpus)
corpusInsert    := corpus.MakeInsert(insertCorpus)
corpusDeleteAll := corpus.MakeDeleteAll(deleteCorpus)
```

### 5. Test with mocks

No real database needed in unit tests — inject mock operations directly:

```go
func TestSelectAll(t *testing.T) {
    expected := []corpus.DAO{{ID: 1, TweetAuthor: "alice"}}
    sel := database.MockSelect(expected, nil)

    selectAll := corpus.MakeSelectAll(sel)
    got, err := selectAll(context.Background())

    assert.NoError(t, err)
    assert.Equal(t, expected, got)
}

func TestInsert_Error(t *testing.T) {
    ins := database.MockInsert[int](-1, errors.New("db error"))

    insert := corpus.MakeInsert(ins)
    _, err := insert(context.Background(), corpus.DTO{})

    assert.Error(t, err)
}

func TestDeleteAll(t *testing.T) {
    del := database.MockDelete(nil)

    deleteAll := corpus.MakeDeleteAll(del)
    err := deleteAll(context.Background())

    assert.NoError(t, err)
}
```

---

## Sentinel errors

All `Make*` helpers return typed sentinel errors on failure:

| Error | Meaning |
|---|---|
| `database.ErrQuery` | `db.Query`, `db.QueryRow`, or `db.Exec` returned an error |
| `database.ErrCollect` | `CollectRows` failed after a successful query |
| `database.ErrNoRows` | `SelectOne` found no matching row (`pgx.ErrNoRows`) |
| `database.ErrAdminNotConfigured` | `OpenAdmin` found no `POSTGRES_ADMIN_USER`/`POSTGRES_ADMIN_PASS` |
| `database.ErrFailedToCreateRole` | `MakeProvision` could not create the role |
| `database.ErrFailedToSyncPassword` | `MakeProvision` could not set the role's password |
| `database.ErrFailedToCreateDatabase` | `MakeProvision` could not create the database |
| `database.ErrFailedToRevokePublic` | `MakeProvision` could not revoke public access to the database |
| `database.ErrFailedToBuildStatement` | `MakeExecFormatted` could not build the statement a query returns; wrapped inside the step's error |
| `database.ErrFailedToExecuteStatement` | `MakeExecFormatted` could not run the statement it built; wrapped inside the step's error |

Use `errors.Is` to handle them in domain code:

```go
result, err := selectOne(ctx, query, id)
if errors.Is(err, database.ErrNoRows) {
    return DAO{}, ErrNotFound
}
```
