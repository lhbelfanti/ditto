package database_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestMakeSelect_success(t *testing.T) {
	mockConn := new(database.MockPostgresConnection)
	mockRows := new(database.MockPgxRows)
	mockConn.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(mockRows, nil)
	collectRows := database.MockCollectRows([]string{"a", "b"}, nil)

	sel := database.MakeSelect[string](mockConn, collectRows)
	got, err := sel(context.Background(), "SELECT name FROM t")

	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b"}, got)
}

func TestMakeSelect_queryError(t *testing.T) {
	mockConn := new(database.MockPostgresConnection)
	mockConn.On("Query", mock.Anything, mock.Anything, mock.Anything).
		Return((*database.MockPgxRows)(nil), pgx.ErrTxClosed)
	collectRows := database.MockCollectRows[string](nil, nil)

	sel := database.MakeSelect[string](mockConn, collectRows)
	got, err := sel(context.Background(), "SELECT name FROM t")

	assert.ErrorIs(t, err, database.ErrQuery)
	assert.Nil(t, got)
}

func TestMakeSelect_collectError(t *testing.T) {
	mockConn := new(database.MockPostgresConnection)
	mockRows := new(database.MockPgxRows)
	mockConn.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(mockRows, nil)
	collectRows := database.MockCollectRows[string](nil, pgx.ErrNoRows)

	sel := database.MakeSelect[string](mockConn, collectRows)
	got, err := sel(context.Background(), "SELECT name FROM t")

	assert.ErrorIs(t, err, database.ErrCollect)
	assert.Nil(t, got)
}

func TestMakeSelectOne_success(t *testing.T) {
	mockConn := new(database.MockPostgresConnection)
	mockRows := new(database.MockPgxRows)
	mockConn.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(mockRows, nil)
	mockRows.On("Next").Return(true)
	mockRows.On("Close").Return()
	mockRows.On("Err").Return(nil)
	mockRows.On("Scan", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(0).([]any)
		*(dest[0].(*bool)) = true
	})
	scanBool := func(row pgx.CollectableRow) (bool, error) {
		var v bool
		return v, row.Scan(&v)
	}

	selOne := database.MakeSelectOne[bool](mockConn, scanBool)
	got, err := selOne(context.Background(), "SELECT applied FROM migrations WHERE name = $1", "000_setup.sql")

	assert.NoError(t, err)
	assert.True(t, got)
}

func TestMakeSelectOne_noRows(t *testing.T) {
	mockConn := new(database.MockPostgresConnection)
	mockRows := new(database.MockPgxRows)
	mockConn.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(mockRows, nil)
	mockRows.On("Next").Return(false)
	mockRows.On("Err").Return(nil)
	mockRows.On("Close").Return()
	scanBool := func(row pgx.CollectableRow) (bool, error) {
		var v bool
		return v, row.Scan(&v)
	}

	selOne := database.MakeSelectOne[bool](mockConn, scanBool)
	_, err := selOne(context.Background(), "SELECT applied FROM migrations WHERE name = $1", "000_setup.sql")

	assert.ErrorIs(t, err, database.ErrNoRows)
}

func TestMakeSelectOne_queryError(t *testing.T) {
	mockConn := new(database.MockPostgresConnection)
	mockConn.On("Query", mock.Anything, mock.Anything, mock.Anything).
		Return((*database.MockPgxRows)(nil), pgx.ErrTxClosed)

	selOne := database.MakeSelectOne[bool](mockConn, nil)
	_, err := selOne(context.Background(), "SELECT applied FROM migrations WHERE name = $1", "000_setup.sql")

	assert.ErrorIs(t, err, database.ErrQuery)
}

func TestMakeInsert_success(t *testing.T) {
	mockConn := new(database.MockPostgresConnection)
	mockRow := new(database.MockPgxRow)
	mockConn.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(0).([]any)
		*(dest[0].(*int)) = 42
	})

	insert := database.MakeInsert[int](mockConn)
	got, err := insert(context.Background(), "INSERT INTO t DEFAULT VALUES RETURNING id")

	assert.NoError(t, err)
	assert.Equal(t, 42, got)
}

func TestMakeInsert_error(t *testing.T) {
	mockConn := new(database.MockPostgresConnection)
	mockRow := new(database.MockPgxRow)
	mockConn.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(mockRow)
	mockRow.On("Scan", mock.Anything).Return(pgx.ErrTxClosed)

	insert := database.MakeInsert[int](mockConn)
	_, err := insert(context.Background(), "INSERT INTO t DEFAULT VALUES RETURNING id")

	assert.ErrorIs(t, err, database.ErrQuery)
}

func TestMakeDelete_success(t *testing.T) {
	mockConn := new(database.MockPostgresConnection)
	mockConn.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)

	del := database.MakeDelete(mockConn)
	err := del(context.Background(), "DELETE FROM t WHERE id = $1", 1)

	assert.NoError(t, err)
}

func TestMakeDelete_error(t *testing.T) {
	mockConn := new(database.MockPostgresConnection)
	mockConn.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, pgx.ErrTxClosed)

	del := database.MakeDelete(mockConn)
	err := del(context.Background(), "DELETE FROM t WHERE id = $1", 1)

	assert.ErrorIs(t, err, database.ErrQuery)
}

func TestMakeUpdate_success(t *testing.T) {
	mockConn := new(database.MockPostgresConnection)
	mockConn.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, nil)

	upd := database.MakeUpdate(mockConn)
	err := upd(context.Background(), "UPDATE t SET name = $1 WHERE id = $2", "x", 1)

	assert.NoError(t, err)
}

func TestMakeUpdate_error(t *testing.T) {
	mockConn := new(database.MockPostgresConnection)
	mockConn.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, pgx.ErrTxClosed)

	upd := database.MakeUpdate(mockConn)
	err := upd(context.Background(), "UPDATE t SET name = $1 WHERE id = $2", "x", 1)

	assert.ErrorIs(t, err, database.ErrQuery)
}
