package database_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestMockSelectExpecting_success(t *testing.T) {
	mockArguments := []any{"BTC", new("Binance")}
	mockSlice := []int{1, 2}
	mockSelectOp := database.MockSelectExpecting(mockArguments, mockSlice)

	want := []int{1, 2}
	got, _ := mockSelectOp(context.Background(), "", "BTC", new("Binance"))

	assert.Equal(t, want, got)
}

func TestMockSelectExpecting_successWhenArgumentsDiffer(t *testing.T) {
	mockArguments := []any{"BTC"}
	mockSlice := []int{1, 2}
	mockSelectOp := database.MockSelectExpecting(mockArguments, mockSlice)

	got, _ := mockSelectOp(context.Background(), "", "ETH")

	assert.Nil(t, got)
}

func TestMockSelectOneExpecting_success(t *testing.T) {
	mockArguments := []any{"BTC", new("Binance")}
	mockSelectOneOp := database.MockSelectOneExpecting(mockArguments, 1)

	want := 1
	got, _ := mockSelectOneOp(context.Background(), "", "BTC", new("Binance"))

	assert.Equal(t, want, got)
}

func TestMockSelectOneExpecting_successWhenArgumentsDiffer(t *testing.T) {
	mockArguments := []any{"BTC"}
	mockSelectOneOp := database.MockSelectOneExpecting(mockArguments, new(1))

	got, _ := mockSelectOneOp(context.Background(), "", "ETH")

	assert.Nil(t, got)
}

func TestMockInsertExpecting_success(t *testing.T) {
	mockArguments := []any{"BTC", new("Binance")}
	mockInsertOp := database.MockInsertExpecting(mockArguments, 1)

	want := 1
	got, _ := mockInsertOp(context.Background(), "", "BTC", new("Binance"))

	assert.Equal(t, want, got)
}

func TestMockInsertExpecting_successWhenArgumentsDiffer(t *testing.T) {
	mockArguments := []any{"BTC"}
	mockInsertOp := database.MockInsertExpecting(mockArguments, 1)

	want := 0
	got, _ := mockInsertOp(context.Background(), "", "ETH")

	assert.Equal(t, want, got)
}

func TestMockDeleteExpecting_success(t *testing.T) {
	mockArguments := []any{1}
	mockDeleteOp := database.MockDeleteExpecting(mockArguments)

	got := mockDeleteOp(context.Background(), "", 1)

	assert.NoError(t, got)
}

func TestMockDeleteExpecting_failsWhenArgumentsDiffer(t *testing.T) {
	mockArguments := []any{1}
	mockDeleteOp := database.MockDeleteExpecting(mockArguments)

	want := database.ErrUnexpectedArguments
	got := mockDeleteOp(context.Background(), "", 2)

	assert.ErrorIs(t, got, want)
}

func TestMockUpdateExpecting_success(t *testing.T) {
	mockArguments := []any{"BTC", 1}
	mockUpdateOp := database.MockUpdateExpecting(mockArguments)

	got := mockUpdateOp(context.Background(), "", "BTC", 1)

	assert.NoError(t, got)
}

func TestMockUpdateExpecting_failsWhenArgumentsDiffer(t *testing.T) {
	mockArguments := []any{"BTC", 1}
	mockUpdateOp := database.MockUpdateExpecting(mockArguments)

	want := database.ErrUnexpectedArguments
	got := mockUpdateOp(context.Background(), "", "BTC")

	assert.ErrorIs(t, got, want)
}

func TestMockSelect_success(t *testing.T) {
	mockSlice := []int{1, 2}
	mockSelectOp := database.MockSelect(mockSlice, nil)

	want := []int{1, 2}
	got, _ := mockSelectOp(context.Background(), "")

	assert.Equal(t, want, got)
}

func TestMockInsert_success(t *testing.T) {
	mockInsertOp := database.MockInsert(1, nil)

	want := 1
	got, _ := mockInsertOp(context.Background(), "")

	assert.Equal(t, want, got)
}

func TestMockInsertCounting_success(t *testing.T) {
	gotCalls := 0
	mockInsertOp := database.MockInsertCounting(1, nil, &gotCalls)

	want := 1
	wantCalls := 1
	got, _ := mockInsertOp(context.Background(), "")

	assert.Equal(t, want, got)
	assert.Equal(t, wantCalls, gotCalls)
}

func TestMockDelete_failsWhenConfiguredToFail(t *testing.T) {
	mockDeleteOp := database.MockDelete(pgx.ErrTxClosed)

	want := pgx.ErrTxClosed
	got := mockDeleteOp(context.Background(), "")

	assert.ErrorIs(t, got, want)
}

func TestMockPostgresConnection_Begin_failsWhenConfiguredToFail(t *testing.T) {
	mockConn := database.MockBeginConnection(nil, pgx.ErrTxClosed)

	want := pgx.ErrTxClosed
	_, got := mockConn.Begin(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMockPgxRows_CommandTag_success(t *testing.T) {
	mockRows := database.MockFailingRows(pgx.ErrTxClosed)

	want := pgconn.CommandTag{}
	got := mockRows.CommandTag()

	assert.Equal(t, want, got)
}

func TestMockPgxRows_FieldDescriptions_success(t *testing.T) {
	mockRows := database.MockFailingRows(pgx.ErrTxClosed)

	got := mockRows.FieldDescriptions()

	assert.Nil(t, got)
}

func TestMockPgxRows_Values_failsWhenRowsFail(t *testing.T) {
	mockRows := database.MockFailingRows(pgx.ErrTxClosed)

	want := pgx.ErrTxClosed
	_, got := mockRows.Values()

	assert.ErrorIs(t, got, want)
}

func TestMockPgxRows_RawValues_success(t *testing.T) {
	mockRows := database.MockFailingRows(pgx.ErrTxClosed)

	got := mockRows.RawValues()

	assert.Nil(t, got)
}

func TestMockPgxRows_Conn_success(t *testing.T) {
	mockRows := database.MockFailingRows(pgx.ErrTxClosed)

	got := mockRows.Conn()

	assert.Nil(t, got)
}

func TestMockPgxTx_Begin_failsWhenTxFails(t *testing.T) {
	mockTx := database.MockTxFailing(pgx.ErrTxClosed)

	want := pgx.ErrTxClosed
	_, got := mockTx.Begin(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMockPgxTx_Commit_failsWhenTxFails(t *testing.T) {
	mockTx := database.MockTxFailing(pgx.ErrTxClosed)

	want := pgx.ErrTxClosed
	got := mockTx.Commit(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMockPgxTx_Rollback_failsWhenTxFails(t *testing.T) {
	mockTx := database.MockTxFailing(pgx.ErrTxClosed)

	want := pgx.ErrTxClosed
	got := mockTx.Rollback(context.Background())

	assert.ErrorIs(t, got, want)
}

func TestMockPgxTx_CopyFrom_failsWhenTxFails(t *testing.T) {
	mockTx := database.MockTxFailing(pgx.ErrTxClosed)

	want := pgx.ErrTxClosed
	_, got := mockTx.CopyFrom(context.Background(), pgx.Identifier{"t"}, []string{"a"}, pgx.CopyFromRows(nil))

	assert.ErrorIs(t, got, want)
}

func TestMockPgxTx_SendBatch_success(t *testing.T) {
	mockTx := database.MockTxFailing(pgx.ErrTxClosed)

	got := mockTx.SendBatch(context.Background(), &pgx.Batch{})

	assert.Nil(t, got)
}

func TestMockPgxTx_LargeObjects_success(t *testing.T) {
	mockTx := database.MockTxFailing(pgx.ErrTxClosed)

	want := pgx.LargeObjects{}
	got := mockTx.LargeObjects()

	assert.Equal(t, want, got)
}

func TestMockPgxTx_Prepare_failsWhenTxFails(t *testing.T) {
	mockTx := database.MockTxFailing(pgx.ErrTxClosed)

	want := pgx.ErrTxClosed
	_, got := mockTx.Prepare(context.Background(), "statement", "SELECT 1")

	assert.ErrorIs(t, got, want)
}

func TestMockPgxTx_Exec_failsWhenTxFails(t *testing.T) {
	mockTx := database.MockTxFailing(pgx.ErrTxClosed)

	want := pgx.ErrTxClosed
	_, got := mockTx.Exec(context.Background(), "SELECT 1")

	assert.ErrorIs(t, got, want)
}

func TestMockPgxTx_Query_failsWhenTxFails(t *testing.T) {
	mockTx := database.MockTxFailing(pgx.ErrTxClosed)

	want := pgx.ErrTxClosed
	_, got := mockTx.Query(context.Background(), "SELECT 1")

	assert.ErrorIs(t, got, want)
}

func TestMockPgxTx_QueryRow_failsWhenTxFails(t *testing.T) {
	mockTx := database.MockTxFailing(pgx.ErrTxClosed)

	want := pgx.ErrTxClosed
	got := mockTx.QueryRow(context.Background(), "SELECT 1").Scan()

	assert.ErrorIs(t, got, want)
}

func TestMockPgxTx_Conn_success(t *testing.T) {
	mockTx := database.MockTxFailing(pgx.ErrTxClosed)

	got := mockTx.Conn()

	assert.Nil(t, got)
}
