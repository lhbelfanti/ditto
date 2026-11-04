package database_test

import (
	"context"
	"testing"

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
