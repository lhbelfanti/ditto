# verification

Proves the journeys of a service end to end, against a real instance of it. A journey is one use case of a
feature card, often across several routes: sign in, revoke a session, check that the session no longer works.
The `verify-<app>` skill of each service launches the instance; this package drives it and reports what happened.

## A journey

A journey is a `Prove` function with the ID its feature card gives it. It builds its requests with ditto's own
request, and calls the methods of `*Run`:

- `NewRequest(method, url)` creates the request, a ditto `*http.Request`, with `WithHeader`, `WithQueryParam`
  and `WithBody` as usual. `URL(path)` puts the base URL of the instance before a path that starts with `/`,
  and leaves the full URL of another service, such as `http://google-mock/authorize`, as it is.
- `Do(request, status)` sends a request as a step of the journey. The step fails when the answer is not
  `status`.
- `Setup(request, status)` sends a request that only prepares the journey, such as a sign-in. Its step is
  marked `Setup`, and it also fails when the answer is not `status`.
- `Decode` reads a JSON body, and `Require` checks a condition on it.

```go
// Revoke proves that a person can end one of their sessions.
func Revoke(run *verification.Run) {
	ana := signIn(run, "ana@example.com")
	second := signIn(run, "ana@example.com")

	list := run.NewRequest(http.MethodGet, run.URL("/accounts/me/sessions/v1"))
	list.WithHeader("Authorization", "Bearer "+ana)

	var sessions SessionsDTO
	run.Decode(run.Do(list, http.StatusOK), &sessions)
	run.Require(len(sessions.Sessions) == 2, "the second sign-in is not listed")

	revoke := run.NewRequest(http.MethodDelete, run.URL("/accounts/me/sessions/"+sessions.Sessions[1].ID+"/v1"))
	revoke.WithHeader("Authorization", "Bearer "+ana)
	run.Do(revoke, http.StatusOK)

	current := run.NewRequest(http.MethodGet, run.URL("/sessions/current/v1"))
	current.WithHeader("Authorization", "Bearer "+second)
	run.Do(current, http.StatusUnauthorized)
}
```

The first failed step stops the journey and records why, as an error that wraps one of the sentinels of
`errors.go`. A journey that panics, for example on an index out of range after a body it did not expect,
fails the same way. The next journey still runs.

## Driving the journeys

`MakeDrive` proves the journeys in order and returns a `Report` with the steps of each one.

```go
drive := verification.MakeDrive(verification.NewClient(10*time.Second), baseURL)
report := drive(ctx, []verification.Journey{{ID: "session.revoke", Prove: Revoke}})

err := verification.WriteReport(evidenceDir, report)
fmt.Print(report.Summary())
if report.Failed() {
	os.Exit(1)
}
```

`NewClient` does not follow redirects, so a journey sees the `302` and its `Location`.

## Evidence

Each journey sends through its own copy of the client, whose transport records what each request carried and
what came back. A step keeps that exchange: its method, its URL with the query, its headers and its JSON
bodies, as the connection saw them. Bearer tokens, cookies, and the `token`, `code`, `state`, `nonce`,
`password` and `secret` fields and query params, among others, are replaced with `<redacted>`, also in a
`Location` header and in nested objects. A body that is not JSON is left out, since a redirect page can repeat
the values of its URL. `WriteReport` saves the report as `report.json`, with each failure as its message.
