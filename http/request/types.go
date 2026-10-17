package request

type (
	// Validator is a request body that checks the presence of its own fields.
	Validator interface {
		IsValid() error
	}

	// lookupParam returns the raw value of the param name, or "" when the request does not send it.
	lookupParam func(name string) string
)
