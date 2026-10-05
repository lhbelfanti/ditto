package log

type (
	// field represent a key-value tuple that will be added to the context
	field struct {
		Key   string
		Value any
	}

	logCtxKey struct{}
)
