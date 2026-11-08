package verification

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunJourney_success(t *testing.T) {
	mockJourney := MockRequiringJourney("health.ping", true, "unused")
	run := &Run{result: Result{Journey: "health.ping"}}
	done := make(chan struct{})

	go runJourney(run, mockJourney.Prove, done)
	<-done

	want := Result{Journey: "health.ping"}
	got := run.result

	assert.Equal(t, want, got)
}

func TestRunJourney_successWhenTheJourneyPanics(t *testing.T) {
	mockJourney := MockPanickingJourney("health.ping", "boom")
	run := &Run{result: Result{Journey: "health.ping"}}
	done := make(chan struct{})

	go runJourney(run, mockJourney.Prove, done)
	<-done

	want := Result{Journey: "health.ping", Failure: "journey panicked: boom"}
	got := run.result

	assert.Equal(t, want, got)
}
