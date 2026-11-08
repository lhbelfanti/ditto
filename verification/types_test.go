package verification_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/verification"
)

func TestReport_Failed_success(t *testing.T) {
	report := verification.Report{Results: []verification.Result{{Journey: "health.ping"}, {Journey: "session.revoke", Failure: "boom"}}}

	got := report.Failed()

	assert.True(t, got)
}

func TestReport_Failed_successWhenEveryJourneyPasses(t *testing.T) {
	report := verification.Report{Results: []verification.Result{{Journey: "health.ping"}}}

	got := report.Failed()

	assert.False(t, got)
}

func TestReport_Summary_success(t *testing.T) {
	report := verification.Report{Results: []verification.Result{
		{Journey: "health.ping", Steps: []verification.Step{{Route: "GET /ping/v1"}}},
		{Journey: "session.revoke", Failure: "boom"},
	}}

	want := "PASS health.ping (1 steps)\nFAIL session.revoke: boom\n2 journeys, 1 failed\n"
	got := report.Summary()

	assert.Equal(t, want, got)
}
