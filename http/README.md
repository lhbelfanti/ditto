# http

The `http` package covers six concerns for a service's HTTP layer: a standardized JSON response
envelope (`response`), reading and checking a request (`request.Bind`), a thin synchronous client (`CustomClient`), ditto's own system routes
(`RegisterSystemRoutes`), bounded graceful shutdown (`GracefulShutdown`), and three middlewares
(`middleware.Auth`/`CORS`/`RequestID`).

For client tests that need a response sequence (for example, 429 followed by 200), set
`NewClient(0).HTTPClient.Transport` to `MockSequenceRoundTripper` with one `MockHTTPResult` per
request. The transport returns an error if more requests are made than results provided.

## `request.Bind`

```go
func ChangeRoleHandlerV1(changeRole ChangeRoleV1) http.HandlerFunc {
    path := request.Fields{"username": request.String}

    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        req, err := request.Bind[RoleRequestDTOV1](r, path, nil)
        if err != nil {
            response.Send(ctx, w, http.StatusBadRequest, err.Error(), nil, err)
            return
        }

        result, err := changeRole(ctx, req.Path.String("username"), req.Body.Role, req.Body.Reason)
        // ...
    }
}
```

`Bind` reads a request in one call: the path params, the query params and the JSON body, in that
order, and returns the first error.

- **Params.** The handler Maker declares them as two `request.Fields` maps, from the name of each
  param to its type: `String` (trimmed), `Int`, `Bool` or `Time` (RFC 3339). A param is required:
  absent, or equal to its zero value (`""`, `0`, `false`), it fails with `ErrMissingPathParam` or
  `ErrMissingQueryParam`. Mark it `.Optional()` when the business layer accepts it left out or at
  its zero value, and use `Values.Has` to tell the two apart. A value that does not parse fails
  with `ErrInvalidPathParam` or `ErrInvalidQueryParam`. Pass `nil` for a map the handler does not
  need.
- **Body.** It is decoded into the type parameter. A decode failure is `ErrInvalidBody`. When the
  type implements `request.Validator` (`IsValid() error`), with a value or a pointer receiver,
  `Bind` calls it and returns its error as it is. Use `request.NoBody` for a handler that reads no
  body: `Bind` then does not touch it.
- **Reading the values.** `req.Path` and `req.Query` have one getter per type (`String`, `Int`,
  `Bool`, `Time`). A getter called with a name that was not declared, or declared with another
  type, panics: it is a bug in the handler, and its tests catch it.

The params are checked in name order, so with several missing ones the error is always the same.
Every error message names the param (`missing path param: username`), so the handler can answer
the 400 with `err.Error()`.

## `RegisterSystemRoutes`

```go
mux := http.NewServeMux()
dittohttp.RegisterSystemRoutes(mux).
    WithDatabasePing(dittohttp.DatabasePing(pg.MakeCheck(2*time.Second)))
```

`RegisterSystemRoutes` always mounts `GET /ping/v1` (unconditional liveness — never touches the
database). `WithDatabasePing` is opt-in and mounts `GET /database/ping/v1` (readiness). Services
built on `app.Run` get both without wiring them by hand.

Migrations are not exposed over HTTP: `app.Run` applies them at boot.

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
| `ErrDatabaseUnavailable` | `WithDatabasePing`'s handler — paired with `ErrMsgDatabaseUnavailable`, since that string is reused both in `errors.New` and separately as `response.Send`'s `message` argument |
| `request.ErrMissingPathParam`/`ErrInvalidPathParam`/`ErrMissingQueryParam`/`ErrInvalidQueryParam`/`ErrInvalidBody` | `request.Bind` — bare: the handler sends `err.Error()`, which names the param, so no message is reused |
| `middleware.ErrMissingAuthHeader`/`middleware.ErrInvalidToken` | `Auth` — same pairing rule: `ErrMsgInvalidToken` is reused across two failure branches, so it's a named constant; the single-use "authorization header required" message stays inline |

A sentinel only gets a paired `ErrMsg*` constant when its exact message string is needed in more
than one place — see `database/errors.go`'s `SafeError` doc for the same rule applied to a
different package.
