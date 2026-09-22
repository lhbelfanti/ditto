package env

// MockLookup returns a Lookup backed by values, for testing RequireValue/RequirePort callers
// without mutating the process environment.
func MockLookup(values map[string]string) Lookup {
	return func(key string) (string, bool) {
		value, ok := values[key]
		return value, ok
	}
}
