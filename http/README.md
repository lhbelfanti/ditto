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
// dtos.go
type (
    // CreateRequestDTOV1 is what CreateHandlerV1 reads from the request.
    CreateRequestDTOV1 struct {
        Path  CreatePathDTOV1
        Query CreateQueryDTOV1
        Body  CreateBodyDTOV1
    }

    // CreatePathDTOV1 holds the path params of CreateHandlerV1.
    CreatePathDTOV1 struct {
        AccountID int `path:"accountID"`
    }

    // CreateQueryDTOV1 holds the query params of CreateHandlerV1.
    CreateQueryDTOV1 struct {
        Notify *bool `query:"notify"`
    }

    // CreateBodyDTOV1 is the body of CreateHandlerV1.
    CreateBodyDTOV1 struct {
        Amount   int    `json:"amount"`
        Currency string `json:"currency"`
    }
)

// IsValid checks that the body has an amount and a currency.
func (dto CreateBodyDTOV1) IsValid() error {
    if dto.Amount == 0 {
        return ErrMissingAmount
    }

    if dto.Currency == "" {
        return ErrMissingCurrency
    }

    return nil
}

// handlers.go
func CreateHandlerV1(create CreateV1) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx := r.Context()

        req, err := request.Bind[CreateRequestDTOV1](r)
        if err != nil {
            response.Send(ctx, w, http.StatusBadRequest, err.Error(), nil, err)
            return
        }

        result, err := create(ctx, req.Path.AccountID, req.Body.Amount, req.Body.Currency, req.Query.Notify)
        // ...
    }
}
```

`Bind` reads a request into a DTO, one per handler, with up to three fields: `Path`, `Query` and
`Body`. Each one is optional: the DTO declares only the parts its handler reads, and a part it
does not declare is not read. A DTO with only a `Path`, only a `Query`, only a `Body`, any two of
them or the three is valid. A handler that reads nothing from the request does not call `Bind`.
It reads path, query and body in that order, each in field order, and returns the first error.

- **Params.** `Path` and `Query` are named struct types, declared in `dtos.go` like the DTO. Each
  field names its param with a `path:"…"` or `query:"…"` tag and is a `string` (trimmed), an int
  of any size, a `bool` or a `time.Time` (RFC 3339). A field is required: absent, or equal to its
  zero value (`""`, `0`, `false`), it fails with `ErrMissingParam`. Make it a pointer when the
  business layer accepts it left out or at its zero value: it stays `nil` when the param is absent.
  A value that does not parse fails with `ErrInvalidParam`, wrapping `ErrNotAnInt`, `ErrNotABool`
  or `ErrNotATime`. The error is wrapped in `ErrInvalidPath` or `ErrInvalidQuery`.
- **Body.** It is decoded from JSON. A decode failure is `ErrInvalidBody`. When its type implements
  `request.Validator` (`IsValid() error`, with a value or a pointer receiver), `Bind` calls it and
  returns its error as it is.
- **A DTO that does not follow these rules** (not a struct, a param field without its tag or not
  exported, a type that is not supported) fails with `ErrInvalidDTO`. It is a bug in the handler,
  and its first test catches it.

Every param error names the part and the param (`invalid path: missing param: accountID`), so the
handler can answer the 400 with `err.Error()`.

## `RegisterSystemRoutes`

```go
mux := http.NewServeMux()
_http.RegisterSystemRoutes(mux).
    WithDatabasePing(_http.DatabasePing(pg.MakeCheck(2*time.Second)))
```

`RegisterSystemRoutes` always mounts `GET /ping/v1` (unconditional liveness — never touches the
database). `WithDatabasePing` is opt-in and mounts `GET /database/ping/v1` (readiness). Services
built on `app.Run` get both without wiring them by hand.

Migrations are not exposed over HTTP: `app.Run` applies them at boot.

## `GracefulShutdown`

```go
signalCtx, stop := signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM)
defer stop()

log.Info(log.With(ctx, log.Param("port", port)), "starting the server")
setup.Must(_http.GracefulShutdown(signalCtx, server.ListenAndServe, server.Shutdown, 5*time.Second))
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
| `ErrFailedToMarshalBody`/`ErrFailedToCreateRequest`/`ErrFailedToExecuteRequest`/`ErrFailedToReadResponse` | `CustomClient.NewRequest`, each with a single call site — bare, no paired message constant. The old names without the `Err` prefix stay as deprecated aliases |
| `ErrFailedToServe`/`ErrFailedToShutdown` | `GracefulShutdown`: the server failed (other than `http.ErrServerClosed`), or the graceful shutdown failed |
| `ErrCantListen` | `Listen`, wrapping the `GracefulShutdown` error |
| `ErrDatabaseUnavailable` | `WithDatabasePing`'s handler — paired with `ErrMsgDatabaseUnavailable`, since that string is reused both in `errors.New` and separately as `response.Send`'s `message` argument |
| `request.ErrInvalidPath`/`ErrInvalidQuery`/`ErrInvalidBody`/`ErrMissingParam`/`ErrInvalidParam`/`ErrNotAnInt`/`ErrNotABool`/`ErrNotATime`/`ErrInvalidDTO` | `request.Bind` — bare: the handler sends `err.Error()`, which names the part and the param, so no message is reused |
| `middleware.ErrMissingAuthHeader`/`middleware.ErrInvalidToken` | `Auth` — same pairing rule: `ErrMsgMissingAuthHeader` and `ErrMsgInvalidToken` are also the `response.Send` messages, so they are named constants. A failed token lookup is wrapped in `ErrInvalidToken`, so its cause reaches the log |

A sentinel only gets a paired `ErrMsg*` constant when its exact message string is needed in more
than one place — see `database/errors.go`'s `SafeError` doc for the same rule applied to a
different package.
