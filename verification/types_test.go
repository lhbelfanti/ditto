package verification_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/verification"
)

func TestReport_Failed_success(t *testing.T) {
	mockErr := errors.New("boom")
	report := verification.Report{Results: []verification.Result{{Journey: "health.ping"}, {Journey: "session.revoke", Failure: mockErr}}}

	got := report.Failed()

	assert.True(t, got)
}

func TestReport_Failed_successWhenEveryJourneyPasses(t *testing.T) {
	report := verification.Report{Results: []verification.Result{{Journey: "health.ping"}}}

	got := report.Failed()

	assert.False(t, got)
}

func TestReport_Summary_success(t *testing.T) {
	mockErr := errors.New("boom")
	report := verification.Report{Results: []verification.Result{
		{Journey: "health.ping", Steps: []verification.Step{{Want: 200, Got: 200}}},
		{Journey: "session.revoke", Failure: mockErr},
	}}

	want := "PASS health.ping (1 steps)\nFAIL session.revoke: boom\n2 journeys, 1 failed\n"
	got := report.Summary()

	assert.Equal(t, want, got)
}

func TestReport_MarshalJSON_success(t *testing.T) {
	mockErr := errors.New("boom")
	report := verification.Report{Results: []verification.Result{
		{Journey: "health.ping", Steps: []verification.Step{{Want: 200, Got: 200}}},
		{Journey: "session.revoke", Failure: mockErr},
	}}

	want := `{"results":[` +
		`{"journey":"health.ping","steps":[{"route":"","want":200,"got":200,"exchange":{"method":"","url":""}}]},` +
		`{"journey":"session.revoke","steps":null,"failure":"boom"}]}`
	got, _ := report.MarshalJSON()

	assert.Equal(t, want, string(got))
}
