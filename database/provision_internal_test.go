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

func TestExecFormatted_failsWhenScanFails(t *testing.T) {
	want := errors.New("boom")
	admin := MockProvisionConnection(MockRowFailing(want), nil)

	got := execFormatted(context.Background(), admin, createRole, "svc")

	assert.ErrorIs(t, got, want)
}

func TestExecFormatted_failsWhenExecFails(t *testing.T) {
	want := errors.New("boom")
	admin := MockProvisionConnection(MockRowReturning("CREATE ROLE svc LOGIN", t), want)

	got := execFormatted(context.Background(), admin, createRole, "svc")

	assert.ErrorIs(t, got, want)
}
