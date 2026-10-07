package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
	"github.com/lhbelfanti/ditto/v2/log"
)

func TestMakeSelect_success(t *testing.T) {
	mockConn := database.MockQueryConnection(&database.MockPgxRows{}, nil)
	collectRows := database.MockCollectRows([]string{"a", "b"}, nil)

	sel := database.MakeSelect[string](mockConn, collectRows)

	want := []string{"a", "b"}
	got, _ := sel(context.Background(), "SELECT name FROM t")

	assert.Equal(t, want, got)
}

func TestMakeSelect_failsWhenQueryFails(t *testing.T) {
	mockConn := database.MockQueryConnection(nil, pgx.ErrTxClosed)
	collectRows := database.MockCollectRows[string](nil, nil)

	sel := database.MakeSelect[string](mockConn, collectRows)

	want := database.ErrQuery
	_, got := sel(context.Background(), "SELECT name FROM t")

	assert.ErrorIs(t, got, want)
}

func TestMakeSelect_failsWhenCollectFails(t *testing.T) {
	mockConn := database.MockQueryConnection(&database.MockPgxRows{}, nil)
	collectRows := database.MockCollectRows[string](nil, pgx.ErrNoRows)

	sel := database.MakeSelect[string](mockConn, collectRows)

	want := database.ErrCollect
	_, got := sel(context.Background(), "SELECT name FROM t")

	assert.ErrorIs(t, got, want)
}

func TestMakeSelectOne_success(t *testing.T) {
	mockConn := database.MockQueryConnection(database.MockRowsReturning(true, t), nil)

	selOne := database.MakeSelectOne[bool](mockConn, pgx.RowTo[bool])

	got, _ := selOne(context.Background(), "SELECT applied FROM migrations WHERE name = $1", "000_setup.sql")

	assert.True(t, got)
}

func TestMakeSelectOne_failsWhenNoRowsFound(t *testing.T) {
	mockConn := database.MockQueryConnection(database.MockEmptyRows(), nil)

	selOne := database.MakeSelectOne[bool](mockConn, pgx.RowTo[bool])

	want := database.ErrNoRows
	_, got := selOne(context.Background(), "SELECT applied FROM migrations WHERE name = $1", "000_setup.sql")

	assert.ErrorIs(t, got, want)
}

func TestMakeSelectOne_failsWhenQueryFails(t *testing.T) {
	mockConn := database.MockQueryConnection(nil, pgx.ErrTxClosed)

	selOne := database.MakeSelectOne[bool](mockConn, nil)

	want := database.ErrQuery
	_, got := selOne(context.Background(), "SELECT applied FROM migrations WHERE name = $1", "000_setup.sql")

	assert.ErrorIs(t, got, want)
}

func TestMakeInsert_success(t *testing.T) {
	mockConn := database.MockQueryRowConnection(database.MockRowReturning(42, t))

	insert := database.MakeInsert[int](mockConn)

	want := 42
	got, _ := insert(context.Background(), "INSERT INTO t DEFAULT VALUES RETURNING id")

	assert.Equal(t, want, got)
}

func TestMakeInsert_failsWhenQueryFails(t *testing.T) {
	mockConn := database.MockQueryRowConnection(database.MockRowFailing(pgx.ErrTxClosed))

	insert := database.MakeInsert[int](mockConn)

	want := database.ErrQuery
	_, got := insert(context.Background(), "INSERT INTO t DEFAULT VALUES RETURNING id")

	assert.ErrorIs(t, got, want)
}

func TestMakeDelete_success(t *testing.T) {
	mockConn := database.MockExecConnection(nil)

	del := database.MakeDelete(mockConn)

	got := del(context.Background(), "DELETE FROM t WHERE id = $1", 1)

	assert.NoError(t, got)
}

func TestMakeDelete_failsWhenQueryFails(t *testing.T) {
	mockConn := database.MockExecConnection(pgx.ErrTxClosed)

	del := database.MakeDelete(mockConn)

	want := database.ErrQuery
	got := del(context.Background(), "DELETE FROM t WHERE id = $1", 1)

	assert.ErrorIs(t, got, want)
}

func TestMakeUpdate_success(t *testing.T) {
	mockConn := database.MockExecConnection(nil)

	upd := database.MakeUpdate(mockConn)

	got := upd(context.Background(), "UPDATE t SET name = $1 WHERE id = $2", "x", 1)

	assert.NoError(t, got)
}

func TestMakeUpdate_failsWhenQueryFails(t *testing.T) {
	mockConn := database.MockExecConnection(pgx.ErrTxClosed)

	upd := database.MakeUpdate(mockConn)

	want := database.ErrQuery
	got := upd(context.Background(), "UPDATE t SET name = $1 WHERE id = $2", "x", 1)

	assert.ErrorIs(t, got, want)
}

func TestMakeSelect_failsWhenQueryFailsAndKeepsTheCause(t *testing.T) {
	mockConn := database.MockQueryConnection(nil, pgx.ErrTxClosed)
	sel := database.MakeSelect[string](mockConn, database.MockCollectRows[string](nil, nil))

	want := pgx.ErrTxClosed
	_, got := sel(context.Background(), "SELECT name FROM t")

func TestMakeCollectRows_success(t *testing.T) {
	collect := database.MakeCollectRows(pgx.RowTo[string])

	got, _ := collect(database.MockEmptyRows())

	assert.Empty(t, got)
}

func TestMakeCollectRows_successWhenFnIsNil(t *testing.T) {
	collect := database.MakeCollectRows[string](nil)

	got, _ := collect(database.MockEmptyRows())

	assert.Empty(t, got)
}

func TestMakeExecFormatted_success(t *testing.T) {
	execFormatted := database.MakeExecFormatted(database.MockSelectOne("CREATE ROLE svc LOGIN", nil), database.MockUpdate(nil))

	got := execFormatted(context.Background(), "SELECT 1")

	assert.NoError(t, got)
}

func TestMakeExecFormatted_successWhenQueryReturnsNoRow(t *testing.T) {
	execFormatted := database.MakeExecFormatted(database.MockSelectOne("", database.ErrNoRows), database.MockUpdate(errors.New("must not run")))

	got := execFormatted(context.Background(), "SELECT 1")

	assert.NoError(t, got)
}

func TestMakeExecFormatted_failsWhenBuildingTheStatementFails(t *testing.T) {
	execFormatted := database.MakeExecFormatted(database.MockSelectOne("", database.ErrQuery), database.MockUpdate(nil))

	want := database.ErrFailedToBuildStatement
	got := execFormatted(context.Background(), "SELECT 1")

	assert.ErrorIs(t, got, want)
}

func TestMakeSelectOne_failsWhenNoRowsFoundAndKeepsTheCause(t *testing.T) {
	mockConn := database.MockQueryConnection(database.MockEmptyRows(), nil)
	selOne := database.MakeSelectOne[string](mockConn, pgx.RowTo[string])

	want := pgx.ErrNoRows
	_, got := selOne(context.Background(), "SELECT name FROM t")

	assert.ErrorIs(t, got, want)
}

func TestMakeUpdate_failsWhenExecFailsAndDoesNotLog(t *testing.T) {
	output := log.MockLogOutput(t)
	update := database.MakeUpdate(database.MockExecConnection(pgx.ErrTxClosed))

	_ = update(context.Background(), "UPDATE t SET a = 1")

	assert.Empty(t, output.String())
}

func TestMakeExecFormatted_failsWhenExecutingTheStatementFails(t *testing.T) {
	execFormatted := database.MakeExecFormatted(database.MockSelectOne("CREATE ROLE svc LOGIN", nil), database.MockUpdate(database.ErrQuery))

	want := database.ErrFailedToExecuteStatement
	got := execFormatted(context.Background(), "SELECT 1")

	assert.ErrorIs(t, got, want)
}
