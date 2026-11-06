package env_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/env"
)

func TestRequireValue_successWhenValueHasSurroundingSpaces(t *testing.T) {
	mockEnv := map[string]string{"KEY": "  value  "}
	mockLookup := env.MockLookup(mockEnv)

	want := "value"
	got, _ := env.RequireValue(mockLookup, "KEY")

	assert.Equal(t, want, got)
}

func TestRequireValue_failsWhenMissing(t *testing.T) {
	mockLookup := env.MockLookup(map[string]string{})

	want := &env.KeyError{Key: "KEY", Err: env.ErrMissingKey}
	_, got := env.RequireValue(mockLookup, "KEY")

	assert.Equal(t, want, got)
}

func TestRequireValue_failsWhenBlank(t *testing.T) {
	mockEnv := map[string]string{"KEY": "   "}
	mockLookup := env.MockLookup(mockEnv)

	want := &env.KeyError{Key: "KEY", Err: env.ErrMissingKey}
	_, got := env.RequireValue(mockLookup, "KEY")

	assert.Equal(t, want, got)
}

func TestRequireSecret_successWhenValueHasSurroundingSpaces(t *testing.T) {
	mockEnv := map[string]string{"KEY": "  value  "}
	mockLookup := env.MockLookup(mockEnv)

	want := "  value  "
	got, _ := env.RequireSecret(mockLookup, "KEY")

	assert.Equal(t, want, got)
}

func TestRequireSecret_failsWhenMissing(t *testing.T) {
	mockLookup := env.MockLookup(map[string]string{})

	want := &env.KeyError{Key: "KEY", Err: env.ErrMissingKey}
	_, got := env.RequireSecret(mockLookup, "KEY")

	assert.Equal(t, want, got)
}

func TestRequireSecret_failsWhenBlank(t *testing.T) {
	mockEnv := map[string]string{"KEY": "   "}
	mockLookup := env.MockLookup(mockEnv)

	want := &env.KeyError{Key: "KEY", Err: env.ErrMissingKey}
	_, got := env.RequireSecret(mockLookup, "KEY")

	assert.Equal(t, want, got)
}

func TestRequirePort_success(t *testing.T) {
	mockEnv := map[string]string{"PORT": "4000"}
	mockLookup := env.MockLookup(mockEnv)

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
			mockEnv := map[string]string{"PORT": tt.value}
			mockLookup := env.MockLookup(mockEnv)

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
			mockEnv := map[string]string{"PORT": tt.value}
			mockLookup := env.MockLookup(mockEnv)

			want := &env.KeyError{Key: "PORT", Err: env.ErrInvalidPort}
			_, got := env.RequirePort(mockLookup, "PORT")

			assert.Equal(t, want, got)
		})
	}
}

func TestRequirePort_failsWhenMissing(t *testing.T) {
	mockLookup := env.MockLookup(map[string]string{})

	want := &env.KeyError{Key: "PORT", Err: env.ErrMissingKey}
	_, got := env.RequirePort(mockLookup, "PORT")

	assert.Equal(t, want, got)
}
