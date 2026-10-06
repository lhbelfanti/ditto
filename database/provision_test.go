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

func TestMakeProvision_failsWhenQueryFails(t *testing.T) {
	admin := database.MockProvisionConnection(database.MockRowFailing(errors.New("boom")), nil)
	provision := database.MakeProvision(admin, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToProvision
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeProvision_failsWhenExecFails(t *testing.T) {
	admin := database.MockProvisionConnection(database.MockRowReturning("CREATE ROLE svc LOGIN", t), errors.New("boom"))
	provision := database.MakeProvision(admin, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToProvision
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeProvision_failsWithoutLeakingTheCause(t *testing.T) {
	admin := database.MockProvisionConnection(database.MockRowReturning("CREATE ROLE svc LOGIN", t), errors.New("password=hunter2"))
	provision := database.MakeProvision(admin, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToProvision.Error()
	got := provision(context.Background()).Error()

	assert.Equal(t, want, got)
}

func TestMakeProvision_failsWhenPasswordSyncFails(t *testing.T) {
	admin := database.MockProvisionConnectionSequence(
		database.MockRowFailing(pgx.ErrNoRows),
		database.MockRowFailing(errors.New("boom")),
	)
	provision := database.MakeProvision(admin, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToProvision
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeProvision_failsWhenDatabaseCreationFails(t *testing.T) {
	admin := database.MockProvisionConnectionSequence(
		database.MockRowFailing(pgx.ErrNoRows),
		database.MockRowFailing(pgx.ErrNoRows),
		database.MockRowFailing(errors.New("boom")),
	)
	provision := database.MakeProvision(admin, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToProvision
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMakeProvision_failsWhenRevokeFails(t *testing.T) {
	admin := database.MockProvisionConnectionSequence(
		database.MockRowFailing(pgx.ErrNoRows),
		database.MockRowFailing(pgx.ErrNoRows),
		database.MockRowFailing(pgx.ErrNoRows),
		database.MockRowFailing(errors.New("boom")),
	)
	provision := database.MakeProvision(admin, database.Target{Name: "svc_db", Role: "svc", Pass: "pw"})

	want := database.ErrFailedToProvision
	got := provision(context.Background())

	assert.ErrorIs(t, got, want)
}
