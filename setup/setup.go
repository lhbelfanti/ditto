package setup

import (
	"context"

	"github.com/lhbelfanti/ditto/v2/log"
)

// Init function logs the error and throws a panic if the initialization of a dependency fails
func Init[T any](val T, err error) T {
	if err != nil {
		log.Err(context.Background(), err, "failed to initialize a dependency")
		panic(err)
	}

	return val
}

// Must function logs the error and throws a panic if the initialization of a dependency fails
func Must(err error) {
	if err != nil {
		log.Err(context.Background(), err, "failed to initialize a dependency")
		panic(err)
	}
}
