# verification

Proves the journeys of a service end to end, against a real instance of it. A journey is one use case of a
feature card, often across several routes: sign in, revoke a session, check that the session no longer works.
The `verify-<app>` skill of each service launches the instance; this package drives it and reports what happened.

## A journey

A journey is a `Prove` function with the ID its feature card gives it. It calls the methods of `*Run`:

- `Do` sends a request as a step of the journey. `Route` is the route as the router registers it, with its
  placeholders, and `Params` fills them. The step fails when the answer is not `Status`.
- `Send` sends a request that only prepares the journey, such as a sign-in. Its step is marked `Setup`, and it
  fails only when `Status` is set and the answer is another one.
- `Decode` reads a JSON body, and `Require` checks a condition on it.

```go
// Revoke proves that a person can end one of their sessions.
func Revoke(run *verification.Run) {
	ana := signIn(run, "ana@example.com")
	second := signIn(run, "ana@example.com")

	var sessions SessionsDTO
	response := run.Do(verification.Request{Route: "GET /accounts/me/sessions/v1", Token: ana, Status: http.StatusOK})
	run.Decode(response, &sessions)
	run.Require(len(sessions.Sessions) == 2, "the second sign-in is not listed")

	run.Do(verification.Request{
		Route:  "DELETE /accounts/me/sessions/{id}/v1",
		Params: map[string]string{"id": sessions.Sessions[1].ID},
		Token:  ana,
		Status: http.StatusOK,
	})
	run.Do(verification.Request{Route: "GET /sessions/current/v1", Token: second, Status: http.StatusUnauthorized})
}
```

The first failed step stops the journey and records why, as an error that wraps one of the sentinels of
`errors.go`. A journey that panics, for example on an index out of range after a body it did not expect,
fails the same way. The next journey still runs.

## Driving the journeys

`MakeDrive` proves the journeys in order and returns a `Report` with the steps of each one. Routes that start
with `/` go to the base URL of the instance; a route with a full URL, such as `GET http://google-mock/authorize`,
goes to that service.

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
