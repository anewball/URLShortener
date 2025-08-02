package shortener

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type mockDatabaseConn struct {
	ExecFunc     func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
	QueryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func (m *mockDatabaseConn) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.QueryRowFunc != nil {
		return m.QueryRowFunc(ctx, sql, args...)
	}
	return nil
}

func (m *mockDatabaseConn) Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
	if m.ExecFunc != nil {
		return m.ExecFunc(ctx, sql, arguments...)
	}
	return pgconn.CommandTag{}, nil
}

func (m *mockDatabaseConn) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.QueryFunc != nil {
		return m.QueryFunc(ctx, sql, args...)
	}
	return nil, nil
}

// mockRow is a mock implementation of pgx.Row.
type mockRow struct {
	result string
	err    error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	if len(dest) < 1 {
		return errors.New("Scan destination is empty")
	}

	// Copy the `result` into the provided destination.
	switch d := dest[0].(type) {
	case *string:
		*d = m.result
	default:
		return errors.New("unsupported type for Scan result")
	}
	return nil
}

func (m *mockRow) FielDescription() []pgconn.FieldDescription {
	return nil // Not needed for this test
}

type mockRows struct {
	data       [][]any
	index      int
	err        error
	scanErrPos int // Position in data where Scan should fail
	closed     bool
}

// CommandTag implements pgx.Rows.
func (m *mockRows) CommandTag() pgconn.CommandTag {
	panic("unimplemented")
}

// Conn implements pgx.Rows.
func (m *mockRows) Conn() *pgx.Conn {
	panic("unimplemented")
}

// FieldDescriptions implements pgx.Rows.
func (m *mockRows) FieldDescriptions() []pgconn.FieldDescription {
	panic("unimplemented")
}

// RawValues implements pgx.Rows.
func (m *mockRows) RawValues() [][]byte {
	panic("unimplemented")
}

// Values implements pgx.Rows.
func (m *mockRows) Values() ([]any, error) {
	panic("unimplemented")
}

func (m *mockRows) Next() bool {
	if m.closed {
		return false
	}
	m.index++
	return m.index <= len(m.data)
}

func (m *mockRows) Scan(dest ...any) error {
	if m.closed {
		return errors.New("rows are closed")
	}
	if m.index-1 >= len(m.data) {
		return errors.New("no row data available")
	}

	if m.scanErrPos > 0 && m.index-1 == m.scanErrPos {
		return errors.New("simulated scan error")
	}

	row := m.data[m.index-1]
	for i, v := range row {
		switch d := dest[i].(type) {
		case *string:
			*d = v.(string)
		case *int:
			*d = v.(int)
		default:
			return errors.New("unsupported destination type")
		}
	}
	return nil
}

func (m *mockRows) Err() error {
	return m.err
}

func (m *mockRows) Close() {
	m.closed = true
}
