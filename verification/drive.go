package verification

import (
	"context"

	"github.com/lhbelfanti/ditto/v2/http"
)

// MakeDrive creates a Drive that sends the requests of each journey through client, with the paths of the app
// resolved against baseURL. Each journey gets its own recorder, so that its records keep only its exchanges.
func MakeDrive(client *http.CustomClient, baseURL string) Drive {
	return func(ctx context.Context, entries []Entry) Report {
		results := make([]Result, 0, len(entries))
		for _, entry := range entries {
			recording, recorder := newRecordingClient(client)
			verifier := &Verifier{ctx: ctx, client: recording, recorder: recorder, baseURL: baseURL, result: Result{Journey: entry.ID}}
			done := make(chan struct{})
			go runJourney(verifier, entry.Journey, done)
			<-done
			results = append(results, verifier.result)
		}

		return Report{Results: results}
	}
}
