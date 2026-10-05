package database

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/mock"
)

type (
	// MockPostgresConnection mock implementation of the postgres db
	MockPostgresConnection struct {
		mock.Mock
	}

	// MockPgxRow mock implementation of pgx.Row
	MockPgxRow struct {
		mock.Mock
	}

	// MockPgxRows mock implementation of pgx.Rows
	MockPgxRows struct {
		mock.Mock
	}

	// MockPgxTx mock implementation of pgx.Tx
	MockPgxTx struct {
		mock.Mock
	}

	// MockPgxCollectableRow mock implementation of pgx.CollectableRow
	MockPgxCollectableRow struct {
		mock.Mock
	}
)

// MockPostgresConnection

// Exec returns the configured command tag and error.
func (m *MockPostgresConnection) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

// Query returns the configured rows and error.
func (m *MockPostgresConnection) Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error) {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgx.Rows), args.Error(1)
}

// QueryRow returns the configured row.
func (m *MockPostgresConnection) QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row {
	args := m.Called(ctx, sql, arguments)
	return args.Get(0).(pgx.Row)
}

// Begin returns the configured transaction and error.
func (m *MockPostgresConnection) Begin(ctx context.Context) (pgx.Tx, error) {
	args := m.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

// MockPgxRow

// Scan returns the configured error.
func (m *MockPgxRow) Scan(dest ...any) error {
	args := m.Called(dest)
	return args.Error(0)
}

// MockPgxRows

// Close records the call.
func (m *MockPgxRows) Close() {
	m.Called()
}

// Err returns the configured error.
func (m *MockPgxRows) Err() error {
	args := m.Called()
	return args.Error(0)
}

// CommandTag returns the configured command tag.
func (m *MockPgxRows) CommandTag() pgconn.CommandTag {
	args := m.Called()
	return args.Get(0).(pgconn.CommandTag)
}

// FieldDescriptions returns nil.
func (m *MockPgxRows) FieldDescriptions() []pgconn.FieldDescription {
	args := m.Called()
	return args.Get(0).([]pgconn.FieldDescription)
}

// Next returns the configured value.
func (m *MockPgxRows) Next() bool {
	args := m.Called()
	return args.Bool(0)
}

// Scan returns the configured error and runs any configured side effect.
func (m *MockPgxRows) Scan(dest ...any) error {
	args := m.Called(dest)
	return args.Error(0)
}

// Values returns the configured values and error.
func (m *MockPgxRows) Values() ([]any, error) {
	args := m.Called()
	return args.Get(0).([]any), args.Error(1)
}

// RawValues returns nil.
func (m *MockPgxRows) RawValues() [][]byte {
	args := m.Called()
	return args.Get(0).([][]byte)
}

// Conn returns nil.
func (m *MockPgxRows) Conn() *pgx.Conn {
	args := m.Called()
	return args.Get(0).(*pgx.Conn)
}

// MockPgxTx

// Begin returns the configured transaction and error.
func (t *MockPgxTx) Begin(ctx context.Context) (pgx.Tx, error) {
	args := t.Called(ctx)
	return args.Get(0).(pgx.Tx), args.Error(1)
}

// Commit returns the configured error.
func (t *MockPgxTx) Commit(ctx context.Context) error {
	args := t.Called(ctx)
	return args.Error(0)
}

// Rollback returns the configured error.
func (t *MockPgxTx) Rollback(ctx context.Context) error {
	args := t.Called(ctx)
	return args.Error(0)
}

// CopyFrom returns the configured row count and error.
func (t *MockPgxTx) CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	args := t.Called(ctx, tableName, columnNames, rowSrc)
	return args.Get(0).(int64), args.Error(1)
}

// SendBatch returns the configured batch results.
func (t *MockPgxTx) SendBatch(ctx context.Context, b *pgx.Batch) pgx.BatchResults {
	args := t.Called(ctx, b)
	return args.Get(0).(pgx.BatchResults)
}

// LargeObjects returns the configured large objects handle.
func (t *MockPgxTx) LargeObjects() pgx.LargeObjects {
	args := t.Called()
	return args.Get(0).(pgx.LargeObjects)
}

// Prepare returns the configured statement description and error.
func (t *MockPgxTx) Prepare(ctx context.Context, name, sql string) (*pgconn.StatementDescription, error) {
	args := t.Called(ctx, name, sql)
	return args.Get(0).(*pgconn.StatementDescription), args.Error(1)
}

// Exec returns the configured command tag and error.
func (t *MockPgxTx) Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error) {
	args := t.Called(ctx, sql, arguments)
	return args.Get(0).(pgconn.CommandTag), args.Error(1)
}

// Query returns the configured rows and error.
func (t *MockPgxTx) Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error) {
	args := t.Called(ctx, sql, arguments)
	return args.Get(0).(pgx.Rows), args.Error(1)
}

// QueryRow returns the configured row.
func (t *MockPgxTx) QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row {
	args := t.Called(ctx, sql, arguments)
	return args.Get(0).(pgx.Row)
}

// Conn returns the configured connection.
func (t *MockPgxTx) Conn() *pgx.Conn {
	args := t.Called()
	return args.Get(0).(*pgx.Conn)
}

// MockPgxCollectableRow

// Scan returns the configured error and runs any configured side effect.
func (m *MockPgxCollectableRow) Scan(dest ...any) error {
	args := m.Called(dest)
	return args.Error(0)
}

// FieldDescriptions returns the configured field descriptions.
func (m *MockPgxCollectableRow) FieldDescriptions() []pgconn.FieldDescription {
	args := m.Called()
	return args.Get(0).([]pgconn.FieldDescription)
}

// Values returns the configured values and error.
func (m *MockPgxCollectableRow) Values() ([]any, error) {
	args := m.Called()
	return args.Get(0).([]any), args.Error(1)
}

// RawValues returns the configured raw values.
func (m *MockPgxCollectableRow) RawValues() [][]byte {
	args := m.Called()
	return args.Get(0).([][]byte)
}

// MockScan mocks the "Scan" func
func MockScan(mockPgxRow *MockPgxRow, values []any, t *testing.T) {
	mockPgxRow.On("Scan", mock.Anything).Return(nil).Run(
		func(args mock.Arguments) {
			dest := args.Get(0).([]any)
			if len(dest) != len(values) {
				t.Errorf("Expected %d destination arguments but got %d", len(values), len(dest))
			}
			for i, val := range values {
				parseScanValue(val, dest[i], t)
			}
		},
	)
}

// MockPgxCollectableRowMethods mocks all the methods of a MockPgxCollectableRow
func MockPgxCollectableRowMethods(m *MockPgxCollectableRow, values []any, t *testing.T) {
	m.On("Scan", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(0).([]any)
		if len(dest) != len(values) {
			t.Errorf("Expected %d destination arguments but got %d", len(values), len(dest))
			return
		}

		for i, val := range values {
			if val == nil {
				continue // Let the zero value remain
			}

			parseScanValue(val, dest[i], t)
		}
	})
}

// MockPgxRowsScanValue makes rows.Scan write value into its first destination.
func MockPgxRowsScanValue(rows *MockPgxRows, value any, t *testing.T) {
	rows.On("Scan", mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		dest := args.Get(0).([]any)
		parseScanValue(value, dest[0], t)
	})
}

// parseScanValue assigns a value from `val` to the provided `dest` based on its type, validating supported types.
// It uses `t` for error reporting in tests when the type is unsupported or mismatched.
func parseScanValue(val any, dest any, t *testing.T) {
	switch d := dest.(type) {
	case *int:
		*d = val.(int)
	case **int:
		*d = val.(*int)
	case *string:
		s, ok := val.(*string)
		if ok {
			*d = *s
		} else {
			*d = val.(string)
		}
	case **string:
		*d = val.(*string)
	case *time.Time:
		*d = val.(time.Time)
	case *bool:
		*d = val.(bool)
	case *[]string:
		*d = val.([]string)
	case *pgtype.Text:
		switch v := val.(type) {
		case *string:
			if v != nil {
				d.String = *v
				d.Valid = true
			} else {
				d.Valid = false
			}
		case string:
			d.String = v
			d.Valid = true
		default:
			d.Valid = false
		}
	case *pgtype.Timestamp:
		v, ok := val.(time.Time)
		if ok {
			d.Time = v
			d.Valid = true
		} else {
			d.Valid = false
		}
	case *pgtype.Bool:
		v, ok := val.(bool)
		if ok {
			d.Bool = v
			d.Valid = true
		} else {
			d.Valid = false
		}
	default:
		t.Errorf("Unsupported type %T", d)
	}
}

// MockCollectRows mocks CollectRows function
func MockCollectRows[T any](slice []T, err error) CollectRows[T] {
	return func(rows pgx.Rows) ([]T, error) {
		return slice, err
	}
}

// MockSelect returns a Select[T] that always returns the given slice and error.
func MockSelect[T any](slice []T, err error) Select[T] {
	return func(ctx context.Context, query string, args ...any) ([]T, error) {
		return slice, err
	}
}

// MockSelectOne returns a SelectOne[T] that always returns the given value and error.
func MockSelectOne[T any](val T, err error) SelectOne[T] {
	return func(ctx context.Context, query string, args ...any) (T, error) {
		return val, err
	}
}

// MockInsert returns an Insert[T] that always returns the given value and error.
func MockInsert[T any](val T, err error) Insert[T] {
	return func(ctx context.Context, query string, args ...any) (T, error) {
		return val, err
	}
}

// MockInsertCounting returns an Insert that increments *calls on every invocation and returns val and err.
func MockInsertCounting[T any](val T, err error, calls *int) Insert[T] {
	return func(context.Context, string, ...any) (T, error) {
		(*calls)++
		return val, err
	}
}

// MockDelete returns a Delete that always returns the given error.
func MockDelete(err error) Delete {
	return func(ctx context.Context, query string, args ...any) error {
		return err
	}
}

// MockUpdate returns an Update that always returns the given error.
func MockUpdate(err error) Update {
	return func(ctx context.Context, query string, args ...any) error {
		return err
	}
}

// MockPing returns a Ping that always returns the given error.
func MockPing(err error) Ping {
	return func(context.Context) error {
		return err
	}
}

// MockBlockingPing returns a Ping that blocks until ctx is done, then returns ctx's error, or err
// if ctx carries none.
func MockBlockingPing(err error) Ping {
	return func(ctx context.Context) error {
		<-ctx.Done()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return err
	}
}

// MockPostgres returns a Postgres over a pool that never connects until used, so any query through it fails with a connection error.
func MockPostgres() *Postgres {
	pool, _ := pgxpool.New(context.Background(), "postgresql://mock:mock@127.0.0.1:1/mock?sslmode=disable")
	return &Postgres{db: pool}
}

// MockExecConnection returns a connection whose Exec succeeds when err is nil and fails with err otherwise.
func MockExecConnection(err error) *MockPostgresConnection {
	conn := &MockPostgresConnection{}
	conn.On("Exec", mock.Anything, mock.Anything, mock.Anything).Return(pgconn.CommandTag{}, err)
	return conn
}

// MockQueryConnection returns a connection whose Query returns rows and err.
func MockQueryConnection(rows *MockPgxRows, err error) *MockPostgresConnection {
	conn := &MockPostgresConnection{}
	conn.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(rows, err)
	return conn
}

// MockQueryRowConnection returns a connection whose QueryRow returns row.
func MockQueryRowConnection(row *MockPgxRow) *MockPostgresConnection {
	conn := &MockPostgresConnection{}
	conn.On("QueryRow", mock.Anything, mock.Anything, mock.Anything).Return(row)
	return conn
}

// MockRowsReturning returns rows holding a single row whose Scan writes value.
func MockRowsReturning(value any, t *testing.T) *MockPgxRows {
	rows := &MockPgxRows{}
	rows.On("Next").Return(true)
	rows.On("Close").Return()
	rows.On("Err").Return(nil)
	MockPgxRowsScanValue(rows, value, t)
	return rows
}

// MockEmptyRows returns rows with no row to read.
func MockEmptyRows() *MockPgxRows {
	rows := &MockPgxRows{}
	rows.On("Next").Return(false)
	rows.On("Close").Return()
	rows.On("Err").Return(nil)
	return rows
}

// MockRowReturning returns a row whose Scan writes value.
func MockRowReturning(value any, t *testing.T) *MockPgxRow {
	row := &MockPgxRow{}
	MockScan(row, []any{value}, t)
	return row
}

// MockRowFailing returns a row whose Scan fails with err.
func MockRowFailing(err error) *MockPgxRow {
	row := &MockPgxRow{}
	row.On("Scan", mock.Anything).Return(err)
	return row
}
