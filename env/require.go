package env

import (
	"strconv"
	"strings"
)

// Lookup retrieves the raw value of an environment variable and whether it was set — matching
// os.LookupEnv's signature so callers can pass it directly, or a mock in tests.
type Lookup func(key string) (string, bool)

// KeyError names the offending environment variable key and the validation condition it failed,
// without exposing its value.
type KeyError struct {
	Key string
	Err error
}

func (e *KeyError) Error() string {
	return e.Key + ": " + e.Err.Error()
}

func (e *KeyError) Unwrap() error {
	return e.Err
}

// RequireValue returns the trimmed, non-empty value of key, or a *KeyError wrapping ErrMissingKey.
func RequireValue(lookup Lookup, key string) (string, error) {
	value, ok := lookup(key)
	trimmed := strings.TrimSpace(value)
	if !ok || trimmed == "" {
		return "", &KeyError{Key: key, Err: ErrMissingKey}
	}

	return trimmed, nil
}

// RequirePort returns key's value parsed as a TCP port in the 1-65535 range, or a *KeyError
// wrapping ErrMissingKey or ErrInvalidPort.
func RequirePort(lookup Lookup, key string) (int, error) {
	value, err := RequireValue(lookup, key)
	if err != nil {
		return 0, err
	}

	port, convErr := strconv.Atoi(value)
	if convErr != nil || port < 1 || port > 65535 {
		return 0, &KeyError{Key: key, Err: ErrInvalidPort}
	}

	return port, nil
}
