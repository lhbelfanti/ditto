package env

type (
	// Lookup retrieves the raw value of an environment variable and whether it was set — matching
	// os.LookupEnv's signature so callers can pass it directly, or a mock in tests.
	Lookup func(key string) (string, bool)

	// KeyError names the offending environment variable key and the validation condition it failed,
	// without exposing its value.
	KeyError struct {
		Key string
		Err error
	}
)

// Error renders the offending key and the condition it failed, without its value.
func (e *KeyError) Error() string {
	return e.Key + ": " + e.Err.Error()
}

// Unwrap returns the validation condition the key failed.
func (e *KeyError) Unwrap() error {
	return e.Err
}
