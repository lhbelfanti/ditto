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
| `CollectRows[T]` | Row scanner passed to `Select[T]` | `pgx.CollectRows` |

---

## End-to-end example

### 1. Initialize the connection

```go
pg, err := database.InitPostgres()
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

Use `errors.Is` to handle them in domain code:

```go
result, err := selectOne(ctx, query, id)
if errors.Is(err, database.ErrNoRows) {
    return DAO{}, ErrNotFound
}
```
