package env_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/env"
)

func TestRequireValue_returnsTrimmedValue(t *testing.T) {
	lookup := env.MockLookup(map[string]string{"KEY": "  value  "})

	got, err := env.RequireValue(lookup, "KEY")

	assert.NoError(t, err)
	assert.Equal(t, "value", got)
}

func TestRequireValue_failsWhenMissing(t *testing.T) {
	lookup := env.MockLookup(map[string]string{})

	_, err := env.RequireValue(lookup, "KEY")

	assert.ErrorIs(t, err, env.ErrMissingKey)
	assert.Equal(t, "KEY: "+env.ErrMissingKey.Error(), err.Error())
}

func TestRequireValue_failsWhenBlank(t *testing.T) {
	lookup := env.MockLookup(map[string]string{"KEY": "   "})

	_, err := env.RequireValue(lookup, "KEY")

	assert.ErrorIs(t, err, env.ErrMissingKey)
}

func TestRequirePort_success(t *testing.T) {
	lookup := env.MockLookup(map[string]string{"PORT": "4000"})

	got, err := env.RequirePort(lookup, "PORT")

	assert.NoError(t, err)
	assert.Equal(t, 4000, got)
}

func TestRequirePort_boundaryValues(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  int
	}{
		{name: "Lowest", value: "1", want: 1},
		{name: "Highest", value: "65535", want: 65535},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookup := env.MockLookup(map[string]string{"PORT": tt.value})

			got, err := env.RequirePort(lookup, "PORT")

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestRequirePort_failsWhenInvalid(t *testing.T) {
	tests := []struct {
		name  string
		value string
	}{
		{name: "NonNumeric", value: "not-a-port"},
		{name: "Zero", value: "0"},
		{name: "Negative", value: "-1"},
		{name: "TooLarge", value: "65536"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lookup := env.MockLookup(map[string]string{"PORT": tt.value})

			_, err := env.RequirePort(lookup, "PORT")

			assert.ErrorIs(t, err, env.ErrInvalidPort)
		})
	}
}

func TestRequirePort_failsWhenMissing(t *testing.T) {
	lookup := env.MockLookup(map[string]string{})

	_, err := env.RequirePort(lookup, "PORT")

	assert.ErrorIs(t, err, env.ErrMissingKey)
}

func TestRequireValue_errorNeverContainsSuppliedValue(t *testing.T) {
	lookup := env.MockLookup(map[string]string{"PORT": "not-a-port"})

	_, err := env.RequirePort(lookup, "PORT")

	assert.NotContains(t, err.Error(), "not-a-port")
}
