package env_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/env"
)

func TestRequireValue_successWhenValueHasSurroundingSpaces(t *testing.T) {
	mockLookup := env.MockLookup(map[string]string{"KEY": "  value  "})

	want := "value"
	got, _ := env.RequireValue(mockLookup, "KEY")

	assert.Equal(t, want, got)
}

func TestRequireValue_failsWhenMissing(t *testing.T) {
	mockLookup := env.MockLookup(map[string]string{})

	want := env.ErrMissingKey
	_, got := env.RequireValue(mockLookup, "KEY")

	assert.ErrorIs(t, got, want)
}

func TestRequireValue_failsWhenMissingNamesTheKey(t *testing.T) {
	mockLookup := env.MockLookup(map[string]string{})

	_, err := env.RequireValue(mockLookup, "KEY")

	want := "KEY: " + env.ErrMissingKey.Error()
	got := err.Error()

	assert.Equal(t, want, got)
}

func TestRequireValue_failsWhenBlank(t *testing.T) {
	mockLookup := env.MockLookup(map[string]string{"KEY": "   "})

	want := env.ErrMissingKey
	_, got := env.RequireValue(mockLookup, "KEY")

	assert.ErrorIs(t, got, want)
}

func TestRequirePort_success(t *testing.T) {
	mockLookup := env.MockLookup(map[string]string{"PORT": "4000"})

	want := 4000
	got, _ := env.RequirePort(mockLookup, "PORT")

	assert.Equal(t, want, got)
}

func TestRequirePort_successWhenPortIsAtTheBoundaries(t *testing.T) {
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
			mockLookup := env.MockLookup(map[string]string{"PORT": tt.value})

			got, _ := env.RequirePort(mockLookup, "PORT")

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
			mockLookup := env.MockLookup(map[string]string{"PORT": tt.value})

			want := env.ErrInvalidPort
			_, got := env.RequirePort(mockLookup, "PORT")

			assert.ErrorIs(t, got, want)
		})
	}
}

func TestRequirePort_failsWhenMissing(t *testing.T) {
	mockLookup := env.MockLookup(map[string]string{})

	want := env.ErrMissingKey
	_, got := env.RequirePort(mockLookup, "PORT")

	assert.ErrorIs(t, got, want)
}

func TestRequirePort_failsWhenValueIsInvalidKeepsValueOutOfError(t *testing.T) {
	mockLookup := env.MockLookup(map[string]string{"PORT": "not-a-port"})

	_, err := env.RequirePort(mockLookup, "PORT")

	assert.NotContains(t, err.Error(), "not-a-port")
}
