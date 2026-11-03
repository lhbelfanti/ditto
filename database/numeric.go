package database

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

// ParseNumeric reads text as a finite PostgreSQL NUMERIC. It rejects text that is not a number and
// the NaN, Infinity and -Infinity literals that a NUMERIC column otherwise accepts.
func ParseNumeric(text string) (pgtype.Numeric, error) {
	var numeric pgtype.Numeric
	err := numeric.Scan(text)
	if err != nil {
		return pgtype.Numeric{}, fmt.Errorf("%w: %w", ErrInvalidNumeric, err)
	}

	if numeric.Int == nil {
		return pgtype.Numeric{}, ErrInvalidNumeric
	}

	return numeric, nil
}
