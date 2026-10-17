package request

type kind int

const (
	kindString kind = iota
	kindInt
	kindBool
	kindTime
)

type (
	// Field describes a path or query param: the type it is parsed into and whether it can be left out.
	Field struct {
		kind     kind
		optional bool
	}

	// Fields maps the name of each path or query param a handler reads to its Field.
	Fields map[string]Field

	// NoBody is the body type of a handler that reads no request body.
	NoBody struct{}

	// Validator is a request body that checks the presence of its own fields.
	Validator interface {
		IsValid() error
	}

	// Values holds the path or query params that Bind parsed, read by name with the getter of their type.
	Values struct {
		fields        Fields
		values        map[string]any
		errUndeclared error
	}

	// Request holds what Bind read from an *http.Request.
	Request[B any] struct {
		Body  B
		Path  Values
		Query Values
	}
)

var (
	// String is a param kept as text, trimmed.
	String = Field{kind: kindString}

	// Int is a param parsed with strconv.Atoi.
	Int = Field{kind: kindInt}

	// Bool is a param parsed with strconv.ParseBool.
	Bool = Field{kind: kindBool}

	// Time is a param parsed as RFC 3339.
	Time = Field{kind: kindTime}
)

// Optional returns a copy of the Field that the request can leave out.
func (f Field) Optional() Field {
	f.optional = true

	return f
}

// String returns the name of the kind.
func (k kind) String() string {
	return [...]string{"String", "Int", "Bool", "Time"}[k]
}
