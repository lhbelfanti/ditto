package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestProvision_success(t *testing.T) {
	mockExecFormatted := database.MockExecFormatted(nil)

	provision := database.MakeProvision(mockExecFormatted, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	got := provision(context.Background())

	assert.NoError(t, got)
}

func TestProvision_failsWhenCreatingTheRoleFails(t *testing.T) {
	mockExecFormatted := database.MockExecFormatted(errors.New("boom"))

	provision := database.MakeProvision(mockExecFormatted, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToCreateRole
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestProvision_failsWhenPasswordSyncFails(t *testing.T) {
	mockExecFormatted := database.MockExecFormattedSequence(nil, errors.New("boom"))

	provision := database.MakeProvision(mockExecFormatted, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToSyncPassword
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestProvision_failsWhenCreatingTheDatabaseFails(t *testing.T) {
	mockExecFormatted := database.MockExecFormattedSequence(nil, nil, errors.New("boom"))

	provision := database.MakeProvision(mockExecFormatted, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToCreateDatabase
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestProvision_failsWhenRevokingPublicAccessFails(t *testing.T) {
	mockExecFormatted := database.MockExecFormattedSequence(nil, nil, nil, errors.New("boom"))

	provision := database.MakeProvision(mockExecFormatted, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToRevokePublic
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestProvision_failsWhenExecFormattedFails(t *testing.T) {
	mockErr := errors.New("boom")
	mockExecFormatted := database.MockExecFormatted(mockErr)

	provision := database.MakeProvision(mockExecFormatted, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToCreateRole
	wantCause := mockErr
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestTargetFromEnv_success(t *testing.T) {
	t.Setenv("POSTGRES_DB_NAME", "svc_db")
	t.Setenv("POSTGRES_DB_USER", "svc")
	t.Setenv("POSTGRES_DB_PASS", "pw")

	want := database.Target{Name: "svc_db", Role: "svc", Pass: "pw"}
	got := database.TargetFromEnv()

	assert.Equal(t, want, got)
}
