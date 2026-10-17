package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestMakeProvision_success(t *testing.T) {
	mockExecFormatted := database.MockExecFormatted(nil)

	provision := database.MakeProvision(mockExecFormatted, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	got := provision(context.Background())

	assert.NoError(t, got)
}

func TestMakeProvision_failsWhenCreatingTheRoleFails(t *testing.T) {
	mockExecFormatted := database.MockExecFormatted(errors.New("boom"))

	provision := database.MakeProvision(mockExecFormatted, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToCreateRole
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeProvision_failsWhenPasswordSyncFails(t *testing.T) {
	mockExecFormatted := database.MockExecFormattedSequence(nil, errors.New("boom"))
	provision := database.MakeProvision(mockExecFormatted, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToSyncPassword
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeProvision_failsWhenCreatingTheDatabaseFails(t *testing.T) {
	mockExecFormatted := database.MockExecFormattedSequence(nil, nil, errors.New("boom"))
	provision := database.MakeProvision(mockExecFormatted, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToCreateDatabase
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeProvision_failsWhenRevokingPublicAccessFails(t *testing.T) {
	mockExecFormatted := database.MockExecFormattedSequence(nil, nil, nil, errors.New("boom"))
	provision := database.MakeProvision(mockExecFormatted, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToRevokePublic
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeProvision_failsWhenExecFormattedFails(t *testing.T) {
	want := errors.New("boom")
	mockExecFormatted := database.MockExecFormatted(want)

	provision := database.MakeProvision(mockExecFormatted, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}
