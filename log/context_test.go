package log_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/log"
)

func TestParam_success(t *testing.T) {
	want := "key"
	got := log.Param("key", "value").Key

	assert.Equal(t, want, got)
}

func TestParam_successWhenValueIsKept(t *testing.T) {
	want := "value"
	got := log.Param("key", "value").Value

	assert.Equal(t, want, got)
}

func TestBaseParam_success(t *testing.T) {
	want := "key"
	got := log.BaseParam("key", "value").Key

	assert.Equal(t, want, got)
}

func TestBaseParam_successWhenValueIsKept(t *testing.T) {
	want := "value"
	got := log.BaseParam("key", "value").Value

	assert.Equal(t, want, got)
}

func TestWith_successWhenChildAddsFieldsParentDoesNotSeeThem(t *testing.T) {
	buf := log.MockLogOutput(t)
	parent := log.With(context.Background(), log.Param("request_id", "abc"))
	_ = log.With(parent, log.Param("worker", 1))
	log.Info(parent, "from the parent")

	entry := map[string]any{}
	_ = json.Unmarshal(buf.Bytes(), &entry)

	got := entry["worker"]

	assert.Nil(t, got)
}

func TestWith_successWhenDerivedConcurrentlyFromTheSameParent(t *testing.T) {
	parent := log.With(context.Background(), log.Param("request_id", "abc"))
	var wg sync.WaitGroup

	for i := range 8 {
		wg.Go(func() {
			_ = log.With(parent, log.Param("worker", i))
		})
	}

	wg.Wait()

	buf := log.MockLogOutput(t)
	log.Info(parent, "from the parent")

	entry := map[string]any{}
	_ = json.Unmarshal(buf.Bytes(), &entry)

	got := entry["worker"]

	assert.Nil(t, got)
}

func TestNewFieldsError_successWhenContextCarriesFields(t *testing.T) {
	ctx := log.With(context.Background(), log.Param("symbol", "BTC/USDT"))

	want := "boom [symbol=BTC/USDT]"
	got := log.NewFieldsError(ctx, errors.New("boom")).Error()

	assert.Equal(t, want, got)
}

func TestNewFieldsError_successWhenCauseIsStillMatched(t *testing.T) {
	cause := errors.New("boom")
	ctx := log.With(context.Background(), log.Param("symbol", "BTC/USDT"))

	got := log.NewFieldsError(ctx, fmt.Errorf("wrapped: %w", cause))

	assert.ErrorIs(t, got, cause)
}

func TestNewFieldsError_successWhenContextHasNoFields(t *testing.T) {
	want := errors.New("boom")
	got := log.NewFieldsError(context.Background(), want)

	assert.Same(t, want, got)
}

func TestNewFieldsError_successWhenErrorIsNil(t *testing.T) {
	got := log.NewFieldsError(log.With(context.Background(), log.Param("symbol", "BTC/USDT")), nil)

	assert.NoError(t, got)
}

func TestNewFieldsError_successWhenFieldIsABaseParam(t *testing.T) {
	ctx := log.With(context.Background(), log.BaseParam("request_id", "abc"))

	want := errors.New("boom")
	got := log.NewFieldsError(ctx, want)

	assert.Same(t, want, got)
}

func TestNewFieldsError_successWhenOnlyParamsAreIncluded(t *testing.T) {
	ctx := log.With(context.Background(), log.BaseParam("request_id", "abc"))
	ctx = log.With(ctx, log.Param("symbol", "BTC/USDT"))

	want := "boom [symbol=BTC/USDT]"
	got := log.NewFieldsError(ctx, errors.New("boom")).Error()

	assert.Equal(t, want, got)
}

func TestNewFieldsError_successWhenParamComesFromTheParentContext(t *testing.T) {
	parent := log.With(context.Background(), log.Param("symbol", "BTC/USDT"))
	ctx := log.With(parent, log.BaseParam("request_id", "abc"))

	want := "boom [symbol=BTC/USDT]"
	got := log.NewFieldsError(ctx, errors.New("boom")).Error()

	assert.Equal(t, want, got)
}

func TestNewFieldsError_successWhenParamIsOverriddenAsABaseParam(t *testing.T) {
	ctx := log.With(context.Background(), log.Param("symbol", "BTC/USDT"))
	ctx = log.With(ctx, log.BaseParam("symbol", "ETH/USDT"))

	want := errors.New("boom")
	got := log.NewFieldsError(ctx, want)

	assert.Same(t, want, got)
}
