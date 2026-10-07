# log

Structured logging on top of `zerolog`. A context can carry fields (`With` and `Param`) that every
log call made with it includes.

## Where errors are logged

Errors are wrapped with `%w` all the way up, so the error that reaches the edge of the program holds
the whole chain. It is logged **once, at that edge**: an HTTP handler (through `response.Send`, which
logs the error and returns only the message to the client), the entry point of a goroutine, consumer
or job, and the program's startup (`setup.Must` and `setup.Init` log before they panic). Code in
between wraps and returns, and does not log. An error that is handled without being returned is
logged where it is handled, at `Warn`.

## Fields that live below the logging point

`With` returns a new context, so a field added deep in the call chain is not visible to the handler
that finally logs the error. `NewFieldsError(ctx, err)` embeds the fields `ctx` carries into the
error's text, so they travel up with it. Call it in the function that adds the fields with `With`,
on the errors that function returns:

```go
ctx = log.With(ctx, log.Param("symbol", symbol), log.Param("timeframe", timeframe))

watermark, err := selectWatermark(ctx, symbol, timeframe)
if err != nil {
    return log.NewFieldsError(ctx, fmt.Errorf("select watermark: %w", err))
}
```

The error text becomes `select watermark: <cause> [symbol=BTC/USDT timeframe=1h]`, with the fields
in alphabetical order and times in RFC 3339. `errors.Is` and `errors.As` still see the cause through
the `*FieldsError`. It returns `err` unchanged when the context carries no fields.
