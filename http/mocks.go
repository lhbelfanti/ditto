package http

import (
	"context"

	"github.com/stretchr/testify/mock"
)

// MockHTTPClient mocks HTTP client
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) NewRequest(ctx context.Context, method, url string, body interface{}) (Response, error) {
	args := m.Called(ctx, method, url, body)
	return args.Get(0).(Response), args.Error(1)
}

// MockDatabasePing builds a DatabasePing that always returns err.
func MockDatabasePing(err error) DatabasePing {
	return func(context.Context) error { return err }
}

// MockMigrationRunner builds a MigrationRunner that always returns err.
func MockMigrationRunner(err error) MigrationRunner {
	return func(context.Context) error { return err }
}
