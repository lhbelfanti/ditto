package request

// Validator is a request body that checks the presence of its own fields.
type Validator interface {
	IsValid() error
}
