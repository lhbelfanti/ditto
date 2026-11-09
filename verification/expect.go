package verification

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/lhbelfanti/ditto/v2/http"
	"github.com/lhbelfanti/ditto/v2/http/response"
)

// Data returns an Expect that decodes the body Send writes and compares its data, whole, with want.
func Data[T any](want T) Expect {
	return func(answer http.Response) error {
		var body response.Body[T]
		err := json.Unmarshal([]byte(answer.Body), &body)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrCantDecodeBody, err)
		}

		if !reflect.DeepEqual(body.Data, want) {
			return fmt.Errorf("%w: got %+v, want %+v", ErrUnexpectedData, body.Data, want)
		}

		return nil
	}
}

// Check returns an Expect that decodes the body Send writes and passes its data to check, for a value that changes
// on each run or that a later step needs.
func Check[T any](check Checker[T]) Expect {
	return func(answer http.Response) error {
		var body response.Body[T]
		err := json.Unmarshal([]byte(answer.Body), &body)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrCantDecodeBody, err)
		}

		err = check(body.Data)
		if err != nil {
			return fmt.Errorf("%w: %w", ErrCheckFailed, err)
		}

		return nil
	}
}
