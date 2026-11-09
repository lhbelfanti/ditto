package verification_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/lhbelfanti/ditto/v2/http"
	"github.com/lhbelfanti/ditto/v2/verification"
)

func TestData_success(t *testing.T) {
	mockResponse := http.Response{Body: `{"code":200,"message":"ok","data":{"id":42}}`}

	expect := verification.Data(map[string]int{"id": 42})

	got := expect(mockResponse)

	assert.NoError(t, got)
}

func TestData_failsWhenTheDataIsAnother(t *testing.T) {
	mockResponse := http.Response{Body: `{"code":200,"data":{"id":7}}`}

	expect := verification.Data(map[string]int{"id": 42})

	want := verification.ErrUnexpectedData
	got := expect(mockResponse)

	assert.ErrorIs(t, got, want)
}

func TestData_failsWhenTheBodyIsNotJSON(t *testing.T) {
	mockResponse := http.Response{Body: "pong"}

	expect := verification.Data("pong")

	want := verification.ErrCantDecodeBody
	got := expect(mockResponse)

	assert.ErrorIs(t, got, want)
}

func TestCheck_success(t *testing.T) {
	mockResponse := http.Response{Body: `{"code":200,"data":{"id":42}}`}
	mockChecker := verification.MockChecker[map[string]int](nil)

	expect := verification.Check(mockChecker)

	got := expect(mockResponse)

	assert.NoError(t, got)
}

func TestCheck_failsWhenTheCheckerFails(t *testing.T) {
	mockResponse := http.Response{Body: `{"code":200,"data":{"id":42}}`}
	mockChecker := verification.MockChecker[map[string]int](errors.New("boom"))

	expect := verification.Check(mockChecker)

	want := verification.ErrCheckFailed
	got := expect(mockResponse)

	assert.ErrorIs(t, got, want)
}

func TestCheck_failsWhenTheBodyIsNotJSON(t *testing.T) {
	mockResponse := http.Response{Body: "pong"}
	mockChecker := verification.MockChecker[string](nil)

	expect := verification.Check(mockChecker)

	want := verification.ErrCantDecodeBody
	got := expect(mockResponse)

	assert.ErrorIs(t, got, want)
}
