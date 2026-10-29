package request

import (
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParse_success(t *testing.T) {
	var section MockEventsQuery
	query := url.Values{"limit": {"20"}, "from": {"2026-01-02T03:04:05Z"}, "done": {"true"}, "actor": {" Nebula "}}
	actor := "Nebula"

	_ = parse(reflect.ValueOf(&section).Elem(), "query", query.Get)

	want := MockEventsQuery{Limit: 20, From: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC), Done: true, Actor: &actor}
	got := section

	assert.Equal(t, want, got)
}

func TestParse_successWhenOptionalParamIsLeftOut(t *testing.T) {
	var section MockOptionalQuery
	query := url.Values{}

	_ = parse(reflect.ValueOf(&section).Elem(), "query", query.Get)

	want := MockOptionalQuery{}
	got := section

	assert.Equal(t, want, got)
}

func TestParse_failsWhenSectionIsNotAStruct(t *testing.T) {
	var section string
	query := url.Values{}

	want := ErrInvalidDTO
	got := parse(reflect.ValueOf(&section).Elem(), "query", query.Get)

	assert.ErrorIs(t, got, want)
}

func TestParse_failsWhenFieldHasNoTag(t *testing.T) {
	var section MockUntaggedPath
	path := url.Values{"username": {"Nebula"}}

	want := ErrInvalidDTO
	got := parse(reflect.ValueOf(&section).Elem(), "path", path.Get)

	assert.ErrorIs(t, got, want)
}

func TestParse_failsWhenFieldIsNotExported(t *testing.T) {
	var section MockUnexportedPath
	path := url.Values{"username": {"Nebula"}}

	want := ErrInvalidDTO
	got := parse(reflect.ValueOf(&section).Elem(), "path", path.Get)

	assert.ErrorIs(t, got, want)
}

func TestParse_failsWhenParamIsMissing(t *testing.T) {
	var section MockAccountPath
	path := url.Values{"username": {"   "}}

	want := ErrMissingParam
	got := parse(reflect.ValueOf(&section).Elem(), "path", path.Get)

	assert.ErrorIs(t, got, want)
}

func TestParse_failsWhenRequiredParamIsItsZeroValue(t *testing.T) {
	var section MockEventsQuery
	query := url.Values{"limit": {"0"}, "from": {"2026-01-02T03:04:05Z"}, "done": {"true"}}

	want := ErrMissingParam
	got := parse(reflect.ValueOf(&section).Elem(), "query", query.Get)

	assert.ErrorIs(t, got, want)
}

func TestParse_failsWhenParamIsInvalid(t *testing.T) {
	var section MockSessionPath
	path := url.Values{"id": {"abc"}}

	want := ErrInvalidParam
	got := parse(reflect.ValueOf(&section).Elem(), "path", path.Get)

	assert.ErrorIs(t, got, want)
}

func TestSet_success(t *testing.T) {
	var target string
	_ = set(reflect.ValueOf(&target).Elem(), "Nebula")

	want := "Nebula"
	got := target

	assert.Equal(t, want, got)
}

func TestSet_successWhenTargetIsAnInt(t *testing.T) {
	var target int64
	_ = set(reflect.ValueOf(&target).Elem(), "42")

	want := int64(42)
	got := target

	assert.Equal(t, want, got)
}

func TestSet_successWhenTargetIsABool(t *testing.T) {
	var target bool
	_ = set(reflect.ValueOf(&target).Elem(), "true")

	got := target

	assert.True(t, got)
}

func TestSet_successWhenTargetIsATime(t *testing.T) {
	var target time.Time
	_ = set(reflect.ValueOf(&target).Elem(), "2026-01-02T03:04:05Z")

	want := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	got := target

	assert.Equal(t, want, got)
}

func TestSet_failsWhenIntIsInvalid(t *testing.T) {
	var target int8

	want := ErrNotAnInt
	got := set(reflect.ValueOf(&target).Elem(), "300")

	assert.ErrorIs(t, got, want)
}

func TestSet_failsWhenBoolIsInvalid(t *testing.T) {
	var target bool

	want := ErrNotABool
	got := set(reflect.ValueOf(&target).Elem(), "maybe")

	assert.ErrorIs(t, got, want)
}

func TestSet_failsWhenTimeIsInvalid(t *testing.T) {
	var target time.Time

	want := ErrNotATime
	got := set(reflect.ValueOf(&target).Elem(), "yesterday")

	assert.ErrorIs(t, got, want)
}

func TestSet_failsWhenTypeIsNotSupported(t *testing.T) {
	var target MockUnsupportedQuery

	want := ErrInvalidDTO
	got := set(reflect.ValueOf(&target.IDs).Elem(), "1")

	assert.ErrorIs(t, got, want)
}
