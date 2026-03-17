package env

import (
	"fmt"
	"os"
)

// Get returns the value of the environment variable named by key.
// If the variable is not set or empty, fallback is returned.
func Get(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// GetOrPanic returns the value of the environment variable named by key.
// It panics if the variable is not set or empty.
func GetOrPanic(key string) string {
	v := os.Getenv(key)
	if v == "" {
		panic(fmt.Sprintf("required environment variable %q is not set", key))
	}
	return v
}
