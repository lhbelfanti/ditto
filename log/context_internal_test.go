package log

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestWith_success(t *testing.T) {
	ctx := With(context.Background(), Param("key1", "value1"), Param("key2", 123))

	want := map[string]any{"key1": "value1", "key2": 123}
	got := ctx.Value(logCtxKey{}).(map[string]any)

	assert.Equal(t, want, got)
}

func TestWith_successWhenContextAlreadyHasParams(t *testing.T) {
	ctx := With(context.Background(), Param("key1", "value1"), Param("key2", 123))
	ctx = With(ctx, Param("key1", "newValue"))

	want := map[string]any{"key1": "newValue", "key2": 123}
	got := ctx.Value(logCtxKey{}).(map[string]any)

	assert.Equal(t, want, got)
}

func TestWithContextParams_success(t *testing.T) {
	tests := []struct {
		name string
		ctx  context.Context
		want string
	}{
		{name: "NilContext", ctx: nil, want: `{}`},
		{name: "EmptyContext", ctx: context.Background(), want: `{}`},
		{
			name: "SingleTypes",
			ctx:  With(context.Background(), Param("string", "value"), Param("int", 42), Param("float64", 42.5), Param("bool", true)),
			want: `{"string":"value","int":42,"float64":42.5,"bool":true}`,
		},
		{name: "Error", ctx: With(context.Background(), Param("error", errors.New("boom"))), want: `{"error":"boom"}`},
		{
			name: "Slices",
			ctx:  With(context.Background(), Param("strings", []string{"a", "b"}), Param("ints", []int{1, 2}), Param("floats", []float64{1.5}), Param("bytes", []byte("value"))),
			want: `{"strings":["a","b"],"ints":[1,2],"floats":[1.5],"bytes":"value"}`,
		},
		{name: "Time", ctx: With(context.Background(), Param("time", time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))), want: `{"time":"2026-01-02T03:04:05Z"}`},
		{name: "OtherType", ctx: With(context.Background(), Param("map", map[string]int{"a": 1})), want: `{"map":{"a":1}}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := zerolog.New(buf)
			event := logger.Log()

			withContextParams(tt.ctx, event).Send()

			got := buf.String()

			assert.JSONEq(t, tt.want, got)
		})
	}
}

func TestFormatField_successWhenValueIsATime(t *testing.T) {
	want := "2026-01-02T03:04:05Z"
	got := formatField(time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC))

	assert.Equal(t, want, got)
}

func TestFormatField_successWhenValueIsNotATime(t *testing.T) {
	want := "42"
	got := formatField(42)

	assert.Equal(t, want, got)
}
