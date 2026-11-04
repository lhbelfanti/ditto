package database_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestMockSelectExpecting_success(t *testing.T) {
	mockSelectOp := database.MockSelectExpecting([]any{"BTC", new("Binance")}, []int{1, 2})

	want := []int{1, 2}
	got, _ := mockSelectOp(context.Background(), "", "BTC", new("Binance"))

	assert.Equal(t, want, got)
}

func TestMockSelectExpecting_successWhenArgumentsDiffer(t *testing.T) {
	mockSelectOp := database.MockSelectExpecting([]any{"BTC"}, []int{1, 2})

	got, _ := mockSelectOp(context.Background(), "", "ETH")

	assert.Nil(t, got)
}

func TestMockSelectOneExpecting_success(t *testing.T) {
	mockSelectOneOp := database.MockSelectOneExpecting([]any{"BTC", new("Binance")}, 1)

	want := 1
	got, _ := mockSelectOneOp(context.Background(), "", "BTC", new("Binance"))

	assert.Equal(t, want, got)
}

func TestMockSelectOneExpecting_successWhenArgumentsDiffer(t *testing.T) {
	mockSelectOneOp := database.MockSelectOneExpecting([]any{"BTC"}, new(1))

	got, _ := mockSelectOneOp(context.Background(), "", "ETH")

	assert.Nil(t, got)
}

func TestMockInsertExpecting_success(t *testing.T) {
	mockInsertOp := database.MockInsertExpecting([]any{"BTC", new("Binance")}, 1)

	want := 1
	got, _ := mockInsertOp(context.Background(), "", "BTC", new("Binance"))

	assert.Equal(t, want, got)
}

func TestMockInsertExpecting_successWhenArgumentsDiffer(t *testing.T) {
	mockInsertOp := database.MockInsertExpecting([]any{"BTC"}, 1)

	want := 0
	got, _ := mockInsertOp(context.Background(), "", "ETH")

	assert.Equal(t, want, got)
}

func TestMockDeleteExpecting_success(t *testing.T) {
	mockDeleteOp := database.MockDeleteExpecting([]any{1})

	got := mockDeleteOp(context.Background(), "", 1)

	assert.NoError(t, got)
}

func TestMockDeleteExpecting_failsWhenArgumentsDiffer(t *testing.T) {
	mockDeleteOp := database.MockDeleteExpecting([]any{1})

	want := database.ErrUnexpectedArguments
	got := mockDeleteOp(context.Background(), "", 2)

	assert.ErrorIs(t, got, want)
}

func TestMockUpdateExpecting_success(t *testing.T) {
	mockUpdateOp := database.MockUpdateExpecting([]any{"BTC", 1})

	got := mockUpdateOp(context.Background(), "", "BTC", 1)

	assert.NoError(t, got)
}

func TestMockUpdateExpecting_failsWhenArgumentsDiffer(t *testing.T) {
	mockUpdateOp := database.MockUpdateExpecting([]any{"BTC", 1})

	want := database.ErrUnexpectedArguments
	got := mockUpdateOp(context.Background(), "", "BTC")

	assert.ErrorIs(t, got, want)
}
