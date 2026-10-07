package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestMakeProvision_success(t *testing.T) {
	admin := database.MockProvisionConnection(database.MockRowReturning("CREATE ROLE svc LOGIN", t), nil)
	provision := database.MakeProvision(admin, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	got := provision(context.Background())

	assert.NoError(t, got)
}

func TestMakeProvision_successWhenRoleAndDatabaseExist(t *testing.T) {
	admin := database.MockProvisionConnection(database.MockRowFailing(pgx.ErrNoRows), nil)
	provision := database.MakeProvision(admin, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	got := provision(context.Background())

	assert.NoError(t, got)
}

func TestMakeProvision_failsWhenCreatingTheRoleFails(t *testing.T) {
	admin := database.MockProvisionConnection(database.MockRowFailing(errors.New("boom")), nil)
	provision := database.MakeProvision(admin, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToCreateRole
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeProvision_failsWhenPasswordSyncFails(t *testing.T) {
	admin := database.MockProvisionConnectionSequence(
		database.MockRowFailing(pgx.ErrNoRows),
		database.MockRowFailing(errors.New("boom")),
	)
	provision := database.MakeProvision(admin, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToSyncPassword
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeProvision_failsWhenCreatingTheDatabaseFails(t *testing.T) {
	admin := database.MockProvisionConnectionSequence(
		database.MockRowFailing(pgx.ErrNoRows),
		database.MockRowFailing(pgx.ErrNoRows),
		database.MockRowFailing(errors.New("boom")),
	)
	provision := database.MakeProvision(admin, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToCreateDatabase
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeProvision_failsWhenRevokingPublicAccessFails(t *testing.T) {
	admin := database.MockProvisionConnectionSequence(
		database.MockRowFailing(pgx.ErrNoRows),
		database.MockRowFailing(pgx.ErrNoRows),
		database.MockRowFailing(pgx.ErrNoRows),
		database.MockRowFailing(errors.New("boom")),
	)
	provision := database.MakeProvision(admin, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToRevokePublic
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeProvision_failsWithTheCauseWhenAStepFails(t *testing.T) {
	want := errors.New("boom")
	admin := database.MockProvisionConnection(database.MockRowFailing(want), nil)
	provision := database.MakeProvision(admin, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}
