package verification

import (
	"context"

	"github.com/lhbelfanti/ditto/v2/http"
)

// MakeDrive creates a Drive that sends the requests of each journey through client, with the routes of the app
// resolved against baseURL.
func MakeDrive(client http.Client, baseURL string) Drive {
	return func(ctx context.Context, journeys []Journey) Report {
		results := make([]Result, 0, len(journeys))
		for _, journey := range journeys {
			run := &Run{ctx: ctx, client: client, baseURL: baseURL, result: Result{Journey: journey.ID}}
			done := make(chan struct{})
			go runJourney(run, journey.Prove, done)
			<-done
			results = append(results, run.result)
		}

		return Report{Results: results}
	}
}
