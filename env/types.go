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
