package database

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
)

func TestExecFormatted_successWhenQueryReturnsNoRow(t *testing.T) {
	admin := MockProvisionConnection(MockRowFailing(pgx.ErrNoRows), errors.New("must not run"))

	got := execFormatted(context.Background(), admin, createRole, "svc")

	assert.NoError(t, got)
}

func TestExecFormatted_failsWhenBuildingTheStatementFails(t *testing.T) {
	admin := MockProvisionConnection(MockRowFailing(errors.New("boom")), nil)

	want := ErrFailedToBuildStatement
	got := execFormatted(context.Background(), admin, createRole, "svc")

	assert.ErrorIs(t, got, want)
}

func TestExecFormatted_failsWhenExecutingTheStatementFails(t *testing.T) {
	admin := MockProvisionConnection(MockRowReturning("CREATE ROLE svc LOGIN", t), errors.New("boom"))

	want := ErrFailedToExecuteStatement
	got := execFormatted(context.Background(), admin, createRole, "svc")

	assert.ErrorIs(t, got, want)
}
