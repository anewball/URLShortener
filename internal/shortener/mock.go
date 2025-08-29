package shortener

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type mockDatabaseConn struct {
	ExecFunc     func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
	QueryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	CloseFunc    func()
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

func (m *mockDatabaseConn) Close() {
	if m.CloseFunc != nil {
		m.CloseFunc()
	}
}

// mockRow is a mock implementation of pgx.Row.
type mockRow struct {
	result []any
	err    error
}

func (m *mockRow) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	if len(dest) != len(m.result) {
		return fmt.Errorf("scan: destination count %d != column count %d", len(dest), len(m.result))
	}

	for i := range dest {
		v := m.result[i]
		switch d := dest[i].(type) {
		case *string:
			if s, ok := v.(string); ok {
				*d = s
			} else {
				return fmt.Errorf("scan: cannot assign %T to *string", v)
			}
		case *uint64:
			switch x := v.(type) {
			case uint64:
				*d = x
			case int64:
				*d = uint64(x)
			case int:
				*d = uint64(x)
			default:
				return fmt.Errorf("scan: cannot assign %T to *uint64", v)
			}
		case *time.Time:
			if tt, ok := v.(time.Time); ok {
				*d = tt
			} else {
				return fmt.Errorf("scan: cannot assign %T to *time.Time", v)
			}
		case **time.Time: // nullable
			switch x := v.(type) {
			case *time.Time:
				*d = x
			case nil:
				*d = nil
			default:
				return fmt.Errorf("scan: cannot assign %T to **time.Time", v)
			}
		default:
			return fmt.Errorf("scan: unsupported destination type %T", d)
		}
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
	// Allow targeted injection of a scan error for testing.
	if m.scanErrPos > 0 && m.index-1 == m.scanErrPos {
		return errors.New("simulated scan error")
	}

	row := m.data[m.index-1]
	if len(dest) > len(row) {
		return fmt.Errorf("scan: destination count %d exceeds available columns %d", len(dest), len(row))
	}

	for i := range dest {
		v := row[i]
		switch d := dest[i].(type) {
		case *string:
			if s, ok := v.(string); ok {
				*d = s
			} else {
				return fmt.Errorf("scan: cannot assign %T to *string", v)
			}
		case *uint64:
			switch x := v.(type) {
			case uint64:
				*d = x
			case int64:
				*d = uint64(x)
			case int:
				*d = uint64(x)
			default:
				return fmt.Errorf("scan: cannot assign %T to *uint64", v)
			}
		case *time.Time:
			if tt, ok := v.(time.Time); ok {
				*d = tt
			} else {
				return fmt.Errorf("scan: cannot assign %T to *time.Time", v)
			}
		case **time.Time: // nullable
			switch x := v.(type) {
			case *time.Time:
				*d = x
			case nil:
				*d = nil
			default:
				return fmt.Errorf("scan: cannot assign %T to **time.Time", v)
			}
		default:
			return fmt.Errorf("scan: unsupported destination type %T", d)
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
