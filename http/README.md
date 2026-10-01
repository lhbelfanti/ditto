# http

The `http` package covers five concerns for a service's HTTP layer: a standardized JSON response
envelope (`response`), a thin synchronous client (`CustomClient`), ditto's own system routes
(`RegisterSystemRoutes`), bounded graceful shutdown (`GracefulShutdown`), and three middlewares
(`middleware.Auth`/`CORS`/`RequestID`).

For client tests that need a response sequence (for example, 429 followed by 200), set
`NewClient(0).HTTPClient.Transport` to `MockSequenceRoundTripper` with one `MockHTTPResult` per
request. The transport returns an error if more requests are made than results provided.

## `RegisterSystemRoutes`

```go
mux := http.NewServeMux()
dittohttp.RegisterSystemRoutes(mux).
    WithMigrationRunner(dittohttp.MigrationRunner(runMigrations)).
    WithDatabasePing(dittohttp.DatabasePing(pg.MakeCheck(2*time.Second)))
```

`RegisterSystemRoutes` always mounts `GET /ping/v1` (unconditional liveness — never touches the
database). `WithMigrationRunner`/`WithDatabasePing` are opt-in: a service only chains in the ones
it actually needs. See `migration/README.md` for wiring a `MigrationRunner`, and `database`'s
`(*Postgres).MakeCheck` for a `DatabasePing`.

## `GracefulShutdown`

```go
signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
defer stop()

log.Info(ctx, "starting on :"+port)
setup.Must(dittohttp.GracefulShutdown(signalCtx, server.ListenAndServe, server.Shutdown, 5*time.Second))
```

Runs `server.ListenAndServe` until `signalCtx` is cancelled (on `SIGINT`/`SIGTERM`), then calls
`server.Shutdown` bounded by the given timeout, so in-flight requests get a chance to finish
before the process exits. `http.ErrServerClosed` is never treated as a failure.

## Middlewares

```go
mux.Handle("/", middleware.RequestID(middleware.CORS()(middleware.Auth(selectUserIDByToken)(protected))))
```

- **`RequestID`** honors an inbound `X-Request-ID` header, or generates one if absent, and injects
  it into the request's logging context — wrap the *outermost* handler with it so every route,
  including public and system ones, gets request correlation, not just an inner protected mux.
- **`CORS`** reads `CORS_ALLOWED_ORIGIN` as a comma-separated allow-list and reflects the
  request's `Origin` back only when it matches one of them (required once more than one origin is
  configured, since `Access-Control-Allow-Origin` accepts exactly one value, and
  `Access-Control-Allow-Credentials: true` rules out a `*` wildcard). It also sets `Vary: Origin`
  so caches keep responses for different origins separate.
- **`Auth`** validates a `Bearer` token via an injected `SelectUserIDByToken` and injects the
  resulting user ID into the request context (`UserIDFromContext`). Failures answer through
  `response.Send`'s JSON envelope, the same shape every other handler in this package uses.

## Sentinel errors

| Error | Where |
|---|---|
| `FailedToMarshalBody`/`FailedToCreateRequest`/`FailedToExecuteRequest`/`FailedToReadResponse` | `CustomClient.NewRequest`, each with a single call site — bare, no paired message constant |
| `ErrDatabaseUnavailable`/`ErrMigrationsFailed` | `RegisterSystemRoutes`'s handlers — paired with `ErrMsgDatabaseUnavailable`/`ErrMsgMigrationsFailed`, since those strings are reused both in `errors.New` and separately as `response.Send`'s `message` argument |
| `middleware.ErrMissingAuthHeader`/`middleware.ErrInvalidToken` | `Auth` — same pairing rule: `ErrMsgInvalidToken` is reused across two failure branches, so it's a named constant; the single-use "authorization header required" message stays inline |

A sentinel only gets a paired `ErrMsg*` constant when its exact message string is needed in more
than one place — see `database/errors.go`'s `SafeError` doc for the same rule applied to a
different package.
