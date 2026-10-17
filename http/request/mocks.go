package request

import (
	"errors"
	"time"
)

// ErrMockInvalidBody is the error that MockInvalidBody and MockInvalidPointerBody return from IsValid.
var ErrMockInvalidBody = errors.New("mock body is not valid")

type (
	// MockBody is a request body whose IsValid, with a value receiver, always passes.
	MockBody struct {
		Role string `json:"role"`
	}

	// MockInvalidBody is a request body whose IsValid, with a value receiver, always fails with ErrMockInvalidBody.
	MockInvalidBody struct {
		Role string `json:"role"`
	}

	// MockInvalidPointerBody is a request body whose IsValid, with a pointer receiver, always fails with ErrMockInvalidBody.
	MockInvalidPointerBody struct {
		Role string `json:"role"`
	}

	// MockAccountPath is the path of a request that names an account.
	MockAccountPath struct {
		Username string `path:"username"`
	}

	// MockEventsQuery is a query with one param of every supported type, all required except Actor.
	MockEventsQuery struct {
		Limit int       `query:"limit"`
		From  time.Time `query:"from"`
		Done  bool      `query:"done"`
		Actor *string   `query:"actor"`
	}

	// MockOptionalQuery is a query whose params are all optional.
	MockOptionalQuery struct {
		Limit   *int  `query:"limit"`
		Confirm *bool `query:"confirm"`
	}

	// MockSessionPath is the path of a request that names a session by a numeric ID.
	MockSessionPath struct {
		ID int64 `path:"id"`
	}

	// MockDeletionPath is a path with two required params.
	MockDeletionPath struct {
		Username string `path:"username"`
		Planet   string `path:"planet"`
	}

	// MockUntaggedPath is a path whose field has no path tag.
	MockUntaggedPath struct {
		Username string
	}

	// MockUnexportedPath is a path whose tagged field is not exported.
	MockUnexportedPath struct {
		username string `path:"username"`
	}

	// MockUnsupportedQuery is a query whose field has a type that Bind does not parse.
	MockUnsupportedQuery struct {
		IDs []int `query:"ids"`
	}

	// MockRequestDTO is a request DTO with Path, Query and Body.
	MockRequestDTO struct {
		Path  MockAccountPath
		Query MockEventsQuery
		Body  MockBody
	}

	// MockOptionalQueryDTO is a request DTO with only an optional Query.
	MockOptionalQueryDTO struct {
		Query MockOptionalQuery
	}

	// MockSessionDTO is a request DTO with only a Path.
	MockSessionDTO struct {
		Path MockSessionPath
	}

	// MockDeletionDTO is a request DTO with a Path of two required params.
	MockDeletionDTO struct {
		Path MockDeletionPath
	}

	// MockInvalidBodyDTO is a request DTO whose Body fails IsValid with a value receiver.
	MockInvalidBodyDTO struct {
		Body MockInvalidBody
	}

	// MockInvalidPointerBodyDTO is a request DTO whose Body fails IsValid with a pointer receiver.
	MockInvalidPointerBodyDTO struct {
		Body MockInvalidPointerBody
	}

	// MockUntaggedDTO is a request DTO whose Path has a field without its tag.
	MockUntaggedDTO struct {
		Path MockUntaggedPath
	}

	// MockNotAStructPathDTO is a request DTO whose Path is not a struct.
	MockNotAStructPathDTO struct {
		Path string
	}
)

// IsValid returns nil.
func (MockBody) IsValid() error {
	return nil
}

// IsValid returns ErrMockInvalidBody.
func (MockInvalidBody) IsValid() error {
	return ErrMockInvalidBody
}

// IsValid returns ErrMockInvalidBody.
func (*MockInvalidPointerBody) IsValid() error {
	return ErrMockInvalidBody
}
