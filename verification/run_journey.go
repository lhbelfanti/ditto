package verification

import "fmt"

// runJourney runs prove on its own goroutine, so that a failed step can stop it with runtime.Goexit, and records a
// panic as its failure.
func runJourney(run *Run, prove Prove, done chan<- struct{}) {
	defer close(done)
	defer func() {
		recovered := recover()
		if recovered != nil {
			run.result.Failure = fmt.Sprintf("%v: %v", ErrJourneyPanicked, recovered)
		}
	}()

	prove(run)
}
