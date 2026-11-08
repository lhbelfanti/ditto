package database

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
)

func TestParseScanValue_successWhenDestinationIsInt(t *testing.T) {
	var dest int
	parseScanValue(1, &dest, t)

	want := 1
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsIntPointer(t *testing.T) {
	var dest *int
	parseScanValue(new(1), &dest, t)

	want := new(1)
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsString(t *testing.T) {
	var dest string
	parseScanValue("a", &dest, t)

	want := "a"
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsStringAndValueIsStringPointer(t *testing.T) {
	var dest string
	parseScanValue(new("a"), &dest, t)

	want := "a"
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsStringPointer(t *testing.T) {
	var dest *string
	parseScanValue(new("a"), &dest, t)

	want := new("a")
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsTime(t *testing.T) {
	var dest time.Time
	parseScanValue(time.Unix(1, 0), &dest, t)

	want := time.Unix(1, 0)
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsBool(t *testing.T) {
	var dest bool
	parseScanValue(true, &dest, t)

	got := dest

	assert.True(t, got)
}

func TestParseScanValue_successWhenDestinationIsStringSlice(t *testing.T) {
	var dest []string
	parseScanValue([]string{"a", "b"}, &dest, t)

	want := []string{"a", "b"}
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsTextAndValueIsStringPointer(t *testing.T) {
	var dest pgtype.Text
	parseScanValue(new("a"), &dest, t)

	want := pgtype.Text{String: "a", Valid: true}
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsTextAndValueIsNilStringPointer(t *testing.T) {
	var dest pgtype.Text
	parseScanValue((*string)(nil), &dest, t)

	want := pgtype.Text{}
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsTextAndValueIsString(t *testing.T) {
	var dest pgtype.Text
	parseScanValue("a", &dest, t)

	want := pgtype.Text{String: "a", Valid: true}
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsTextAndValueIsNil(t *testing.T) {
	var dest pgtype.Text
	parseScanValue(nil, &dest, t)

	want := pgtype.Text{}
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsTimestamp(t *testing.T) {
	var dest pgtype.Timestamp
	parseScanValue(time.Unix(1, 0), &dest, t)

	want := pgtype.Timestamp{Time: time.Unix(1, 0), Valid: true}
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsTimestampAndValueIsNil(t *testing.T) {
	var dest pgtype.Timestamp
	parseScanValue(nil, &dest, t)

	want := pgtype.Timestamp{}
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsBoolType(t *testing.T) {
	var dest pgtype.Bool
	parseScanValue(true, &dest, t)

	want := pgtype.Bool{Bool: true, Valid: true}
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_successWhenDestinationIsBoolTypeAndValueIsNil(t *testing.T) {
	var dest pgtype.Bool
	parseScanValue(nil, &dest, t)

	want := pgtype.Bool{}
	got := dest

	assert.Equal(t, want, got)
}

func TestParseScanValue_failsWhenDestinationIsUnsupported(t *testing.T) {
	isolatedT := CaptureTestingT()
	var dest float64
	parseScanValue(1.5, &dest, isolatedT)

	got := isolatedT.Failed()

	assert.True(t, got)
}
