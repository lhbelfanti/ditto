package log_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/log"
)

func TestFieldsError_Error_successWhenFieldsAreSorted(t *testing.T) {
	ctx := log.With(context.Background(), log.EmbeddedParam("timeframe", "1h"), log.EmbeddedParam("symbol", "BTC/USDT"))
	err := log.NewFieldsError(ctx, errors.New("boom"))

	want := "boom [symbol=BTC/USDT timeframe=1h]"
	got := err.Error()

	assert.Equal(t, want, got)
}

func TestFieldsError_Error_successWhenValueIsATime(t *testing.T) {
	ctx := log.With(context.Background(), log.EmbeddedParam("from", time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)))
	err := log.NewFieldsError(ctx, errors.New("boom"))

	want := "boom [from=2026-01-02T03:04:05Z]"
	got := err.Error()

	assert.Equal(t, want, got)
}

func TestFieldsError_Unwrap_success(t *testing.T) {
	cause := errors.New("boom")
	ctx := log.With(context.Background(), log.EmbeddedParam("symbol", "BTC/USDT"))
	err := log.NewFieldsError(ctx, cause)

	got := errors.Unwrap(err)

	assert.Equal(t, cause, got)
}
