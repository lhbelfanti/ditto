# verification

Proves the journeys of a service end to end, against a real instance of it. A journey is one use case of a
feature card, often across several routes: sign in, revoke a session, check that the session no longer works.
The `verify-<app>` skill of each service launches the instance; this package drives it and reports what happened.

## A journey

Every journey has the same shape: one file, one type, two methods. The drive runs them in the same order for
all of them, so a journey only says what it needs and what it proves.

- `Setup(verifier)` prepares what the steps need, such as sessions or an account in a given state. It sends
  its requests with `verifier.Send(request, status)`, which stops the journey when the answer is not `status`
  and marks the request as a setup request in the report. `verifier.NewRequest(method, verifier.URL(path))`
  builds them, and `verifier.Decode` reads a JSON body.
- `Steps()` lists the requests the journey proves, in the order they are sent. Each `Step` has a
  `Description`, the request (`Method`, `Path`, `PathValues`, `Query`, `Bearer`, `Body`, `With`), the `Status`
  it must answer and, when the status is not enough, an `Expect`.

The drive calls `Setup`, then sends each step in order and checks its status and its `Expect`. The first step
that fails stops the journey and records why, as an error that wraps one of the sentinels of `errors.go`. A
journey that panics, for example on an index out of range after a body it did not expect, fails the same way.
The next journey still runs.

```go
// RevokeOne proves that a person ends one of their other sessions.
type RevokeOne struct {
	current string
	other   string
	otherID string
}

// Setup signs one person in twice.
func (j *RevokeOne) Setup(verifier *verification.Verifier) {
	j.current = identity.SignIn(verifier, "revoke-one@example.com")
	j.other = identity.SignIn(verifier, "revoke-one@example.com")
}

// Steps lists the sessions, revokes the other one and checks that it no longer works.
func (j *RevokeOne) Steps() []verification.Step {
	return []verification.Step{
		{
			Description: "lists both sessions",
			Method:      http.MethodGet,
			Path:        "/accounts/me/sessions/v1",
			Bearer:      j.current,
			Status:      http.StatusOK,
			Expect:      verification.Check(j.keepOther),
		},
		{
			Description: "revokes the other session",
			Method:      http.MethodDelete,
			Path:        "/accounts/me/sessions/{id}/v1",
			PathValues:  func() map[string]string { return map[string]string{"id": j.otherID} },
			Bearer:      j.current,
			Status:      http.StatusOK,
		},
		{
			Description: "the other session no longer works",
			Method:      http.MethodGet,
			Path:        "/sessions/current/v1",
			Bearer:      j.other,
			Status:      http.StatusUnauthorized,
		},
	}
}
```

`Steps` runs after `Setup`, so a step reads what `Setup` left, such as `j.current`. A value that only an
earlier step knows is read when the step is sent: `PathValues` is a function, and `With` changes the request
before it goes, for example with `request.WithCookies(j.start)` to send back the cookies of an earlier
response. A `Path` that is a full URL, of another service, is sent as it is, so `Path: "{location}"` with the
`Location` of an earlier redirect follows it.

`Expect` has two helpers, both for the body `response.Send` writes:

- `verification.Data(want)` compares its whole `data` with `want`.
- `verification.Check(func(data T) error)` passes its `data` to a function, for a value that changes on each
  run, such as a date, or that a later step needs, such as the ID of a session.

An `Expect` is any `func(response http.Response) error`, for a response that is not JSON.

## Driving the journeys

`MakeDrive` proves the journeys in order and returns a `Report` with the requests of each one.

```go
entries := []verification.Entry{{ID: "sessions.revoke-one", Journey: &sessions.RevokeOne{}}}

drive := verification.MakeDrive(verification.NewClient(10*time.Second), baseURL)
report := drive(ctx, entries)

err := verification.WriteReport(evidenceDir, report)
fmt.Print(report.Summary())
if report.Failed() {
	os.Exit(1)
}
```

`NewClient` does not follow redirects, so a journey sees the `302` and its `Location`.

## Feature cards

A feature card lists each journey under a `### <ID> — <title>` heading, with a table of its steps.
`UpdateCards(dir, entries)` writes that table from `Steps()`, so the card always says what the journey sends:

```markdown
| # | Request | Status | Proves |
|---|---|---|---|
| 1 | `GET /accounts/me/sessions/v1` | 200 | lists both sessions |
| 2 | `DELETE /accounts/me/sessions/{id}/v1` | 200 | revokes the other session |
| 3 | `GET /sessions/current/v1` | 401 | the other session no longer works |
```

It replaces the first table under the heading of each journey, and fails with `ErrCardNotFound` when a journey
has no heading with a table in any card of `dir`. It calls `Steps()` without `Setup`, so `Steps()` only reads
fields and never sends anything. `Table(steps)` returns the table alone. A CI step that runs it and then
`git diff --exit-code` fails when a card and its journey disagree.

## Evidence

Each journey sends through its own copy of the client, whose transport records what each request carried and
what came back. A record keeps that exchange: its method, its URL with the query, its headers and its JSON
bodies, as the connection saw them, and the description of the step it proves. Bearer tokens, cookies, and the
`token`, `code`, `state`, `nonce`, `password` and `secret` fields and query params, among others, are replaced
with `<redacted>`, also in a `Location` header and in nested objects. Only a string value is redacted, so the
numeric `code` of the envelope stays. A body that is not JSON is left out, since a redirect page can repeat the
values of its URL. `WriteReport` saves the report as `report.json`, with each failure as its message.
