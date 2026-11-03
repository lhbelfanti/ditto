package database_test

import (
	"math/big"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/database"
)

func TestParseNumeric_success(t *testing.T) {
	want := pgtype.Numeric{Int: big.NewInt(-15), Exp: -1, Valid: true}
	got, _ := database.ParseNumeric("-1.5")

	assert.Equal(t, want, got)
}

func TestParseNumeric_failsWhenTextIsNotANumber(t *testing.T) {
	want := database.ErrInvalidNumeric
	_, got := database.ParseNumeric("abc")

	assert.ErrorIs(t, got, want)
}

func TestParseNumeric_failsWhenNumericIsNotFinite(t *testing.T) {
	for _, text := range database.MockNonFiniteNumerics() {
		t.Run(text, func(t *testing.T) {
			want := database.ErrInvalidNumeric
			_, got := database.ParseNumeric(text)

			assert.ErrorIs(t, got, want)
		})
	}
}
