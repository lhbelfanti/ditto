package database_test

import (
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestOptionalText_success(t *testing.T) {
	want := new("titan")
	got := database.OptionalText("titan")

	assert.Equal(t, want, got)
}

func TestOptionalText_successWhenEmpty(t *testing.T) {
	got := database.OptionalText("")

	assert.Nil(t, got)
}

func TestOptionalNumeric_success(t *testing.T) {
	var want pgtype.Numeric

	_ = want.Scan("64000.50")

	got, _ := database.OptionalNumeric(new("64000.50"))

	assert.Equal(t, want, got)
}

func TestOptionalNumeric_successWhenNil(t *testing.T) {
	want := pgtype.Numeric{}
	got, _ := database.OptionalNumeric(nil)

	assert.Equal(t, want, got)
}

func TestOptionalNumeric_failsWhenNotANumber(t *testing.T) {
	want := database.ErrInvalidNumeric
	_, got := database.OptionalNumeric(new("abc"))

	assert.ErrorIs(t, got, want)
}
