package database

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

// OptionalText returns nil for an empty value, so the column stores SQL NULL instead of an empty string.
func OptionalText(value string) *string {
	if value == "" {
		return nil
	}

	return &value
}

// OptionalNumeric scans value, exact decimal text, into a pgtype.Numeric that encodes as SQL NULL
// when value is nil, so the stored value matches the submitted digits with no float round-trip.
func OptionalNumeric(value *string) (pgtype.Numeric, error) {
	if value == nil {
		return pgtype.Numeric{}, nil
	}

	var numeric pgtype.Numeric
	err := numeric.Scan(*value)
	if err != nil {
		return pgtype.Numeric{}, fmt.Errorf("%w: %w", ErrInvalidNumeric, err)
	}

	return numeric, nil
}
