package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestSelect_success(t *testing.T) {
	mockConn := database.MockQueryConnection(&database.MockPgxRows{}, nil)
	mockCollectRows := database.MockCollectRows([]string{"a", "b"}, nil)

	sel := database.MakeSelect[string](mockConn, mockCollectRows)

	want := []string{"a", "b"}
	got, _ := sel(context.Background(), "SELECT name FROM t")

	assert.Equal(t, want, got)
}

func TestSelect_failsWhenQueryFails(t *testing.T) {
	mockConn := database.MockQueryConnection(nil, pgx.ErrTxClosed)
	mockCollectRows := database.MockCollectRows[string](nil, nil)

	sel := database.MakeSelect[string](mockConn, mockCollectRows)

	want := database.ErrQuery
	wantCause := pgx.ErrTxClosed
	_, got := sel(context.Background(), "SELECT name FROM t")

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestSelect_failsWhenCollectFails(t *testing.T) {
	mockConn := database.MockQueryConnection(&database.MockPgxRows{}, nil)
	mockCollectRows := database.MockCollectRows[string](nil, pgx.ErrNoRows)

	sel := database.MakeSelect[string](mockConn, mockCollectRows)

	want := database.ErrCollect
	wantCause := pgx.ErrNoRows
	_, got := sel(context.Background(), "SELECT name FROM t")

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestSelectOne_success(t *testing.T) {
	mockConn := database.MockQueryConnection(database.MockRowsReturning(true, t), nil)

	selOne := database.MakeSelectOne[bool](mockConn, pgx.RowTo[bool])

	got, _ := selOne(context.Background(), "SELECT applied FROM migrations WHERE name = $1", "000_setup.sql")

	assert.True(t, got)
}

func TestSelectOne_failsWhenNoRowsFound(t *testing.T) {
	mockConn := database.MockQueryConnection(database.MockEmptyRows(), nil)

	selOne := database.MakeSelectOne[bool](mockConn, pgx.RowTo[bool])

	want := database.ErrNoRows
	wantCause := pgx.ErrNoRows
	_, got := selOne(context.Background(), "SELECT applied FROM migrations WHERE name = $1", "000_setup.sql")

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestSelectOne_failsWhenQueryFails(t *testing.T) {
	mockConn := database.MockQueryConnection(nil, pgx.ErrTxClosed)

	selOne := database.MakeSelectOne[bool](mockConn, nil)

	want := database.ErrQuery
	wantCause := pgx.ErrTxClosed
	_, got := selOne(context.Background(), "SELECT applied FROM migrations WHERE name = $1", "000_setup.sql")

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestInsert_success(t *testing.T) {
	mockConn := database.MockQueryRowConnection(database.MockRowReturning(42, t))

	insert := database.MakeInsert[int](mockConn)

	want := 42
	got, _ := insert(context.Background(), "INSERT INTO t DEFAULT VALUES RETURNING id")

	assert.Equal(t, want, got)
}

func TestInsert_failsWhenQueryFails(t *testing.T) {
	mockConn := database.MockQueryRowConnection(database.MockRowFailing(pgx.ErrTxClosed))

	insert := database.MakeInsert[int](mockConn)

	want := database.ErrQuery
	wantCause := pgx.ErrTxClosed
	_, got := insert(context.Background(), "INSERT INTO t DEFAULT VALUES RETURNING id")

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestDelete_success(t *testing.T) {
	mockConn := database.MockExecConnection(nil)

	del := database.MakeDelete(mockConn)

	got := del(context.Background(), "DELETE FROM t WHERE id = $1", 1)

	assert.NoError(t, got)
}

func TestDelete_failsWhenQueryFails(t *testing.T) {
	mockConn := database.MockExecConnection(pgx.ErrTxClosed)

	del := database.MakeDelete(mockConn)

	want := database.ErrQuery
	wantCause := pgx.ErrTxClosed
	got := del(context.Background(), "DELETE FROM t WHERE id = $1", 1)

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestUpdate_success(t *testing.T) {
	mockConn := database.MockExecConnection(nil)

	upd := database.MakeUpdate(mockConn)

	got := upd(context.Background(), "UPDATE t SET name = $1 WHERE id = $2", "x", 1)

	assert.NoError(t, got)
}

func TestUpdate_failsWhenQueryFails(t *testing.T) {
	mockConn := database.MockExecConnection(pgx.ErrTxClosed)

	upd := database.MakeUpdate(mockConn)

	want := database.ErrQuery
	wantCause := pgx.ErrTxClosed
	got := upd(context.Background(), "UPDATE t SET name = $1 WHERE id = $2", "x", 1)

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestCollectRows_success(t *testing.T) {
	collect := database.MakeCollectRows(pgx.RowTo[string])

	got, _ := collect(database.MockEmptyRows())

	assert.Empty(t, got)
}

func TestCollectRows_successWhenFnIsNil(t *testing.T) {
	collect := database.MakeCollectRows[string](nil)

	got, _ := collect(database.MockEmptyRows())

	assert.Empty(t, got)
}

func TestExecFormatted_success(t *testing.T) {
	mockSelectOneOp := database.MockSelectOne("CREATE ROLE svc LOGIN", nil)
	mockUpdateOp := database.MockUpdate(nil)

	execFormatted := database.MakeExecFormatted(mockSelectOneOp, mockUpdateOp)

	got := execFormatted(context.Background(), "SELECT 1")

	assert.NoError(t, got)
}

func TestExecFormatted_successWhenQueryReturnsNoRow(t *testing.T) {
	mockSelectOneOp := database.MockSelectOne("", database.ErrNoRows)
	mockUpdateOp := database.MockUpdate(errors.New("must not run"))

	execFormatted := database.MakeExecFormatted(mockSelectOneOp, mockUpdateOp)

	got := execFormatted(context.Background(), "SELECT 1")

	assert.NoError(t, got)
}

func TestExecFormatted_failsWhenBuildingTheStatementFails(t *testing.T) {
	mockSelectOneOp := database.MockSelectOne("", database.ErrQuery)
	mockUpdateOp := database.MockUpdate(nil)

	execFormatted := database.MakeExecFormatted(mockSelectOneOp, mockUpdateOp)

	want := database.ErrFailedToBuildStatement
	wantCause := database.ErrQuery
	got := execFormatted(context.Background(), "SELECT 1")

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}

func TestExecFormatted_failsWhenExecutingTheStatementFails(t *testing.T) {
	mockSelectOneOp := database.MockSelectOne("CREATE ROLE svc LOGIN", nil)
	mockUpdateOp := database.MockUpdate(database.ErrQuery)

	execFormatted := database.MakeExecFormatted(mockSelectOneOp, mockUpdateOp)

	want := database.ErrFailedToExecuteStatement
	wantCause := database.ErrQuery
	got := execFormatted(context.Background(), "SELECT 1")

	assert.ErrorIs(t, got, want)
	assert.ErrorIs(t, got, wantCause)
}
