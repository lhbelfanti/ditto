package broker

import "context"

// MockEnqueue returns a mock MessageBroker that can be used in tests.
func MockEnqueue(err error) MessageBroker {
	return &mockBroker{enqueueErr: err}
}

type mockBroker struct {
	enqueueErr error
}

func (m *mockBroker) EnqueueMessage(_ context.Context, _ string) error {
	return m.enqueueErr
}

func (m *mockBroker) InitMessageConsumerWithFunction(_ int, _ ProcessorFunction) {}

func (m *mockBroker) CloseConnection() {}
