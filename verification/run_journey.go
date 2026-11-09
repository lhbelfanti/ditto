package verification

import "fmt"

// runJourney sets journey up and sends its steps in order, on its own goroutine, so that a failed step can stop it
// with runtime.Goexit. It records a panic as the failure of the journey.
func runJourney(verifier *Verifier, journey Journey, done chan<- struct{}) {
	defer close(done)
	defer recordPanic(verifier)

	journey.Setup(verifier)
	for _, step := range journey.Steps() {
		verifier.prove(step)
	}
}

// recordPanic records as the failure of verifier what the deferring journey panicked with.
func recordPanic(verifier *Verifier) {
	recovered := recover()
	if recovered != nil {
		verifier.result.Failure = fmt.Errorf("%w: %v", ErrJourneyPanicked, recovered)
	}
}
