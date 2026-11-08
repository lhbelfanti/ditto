package verification_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/verification"
)

func TestWriteReport_success(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "evidence")
	report := verification.Report{Results: []verification.Result{{Journey: "health.ping"}}}

	err := verification.WriteReport(dir, report)

	assert.NoError(t, err)
}

func TestWriteReport_failsWhenTheDirectoryCantBeCreated(t *testing.T) {
	report := verification.Report{}

	want := verification.ErrCantCreateDir
	got := verification.WriteReport("/dev/null/evidence", report)

	assert.ErrorIs(t, got, want)
}

func TestWriteReport_failsWhenTheReportCantBeWritten(t *testing.T) {
	dir := t.TempDir()
	_ = os.Mkdir(filepath.Join(dir, "report.json"), 0o755)
	report := verification.Report{}

	want := verification.ErrCantWriteReport
	got := verification.WriteReport(dir, report)

	assert.ErrorIs(t, got, want)
}
