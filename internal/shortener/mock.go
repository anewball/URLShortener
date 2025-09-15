package shortener

import (
	"context"
	"fmt"
	"time"

	"github.com/anewball/urlshortener/internal/dbiface"
)

type mockDatabaseConn struct {
	ExecFunc     func(ctx context.Context, sql string, arguments ...any) (dbiface.CommandResult, error)
	QueryRowFunc func(ctx context.Context, sql string, args ...any) dbiface.Row
	QueryFunc    func(ctx context.Context, sql string, args ...any) (dbiface.Rows, error)
	CloseFunc    func()
}

func (m *mockDatabaseConn) QueryRow(ctx context.Context, sql string, args ...any) dbiface.Row {
	return m.QueryRowFunc(ctx, sql, args...)
}

func (m *mockDatabaseConn) Exec(ctx context.Context, sql string, arguments ...any) (dbiface.CommandResult, error) {
	return m.ExecFunc(ctx, sql, arguments...)
}

func (m *mockDatabaseConn) Query(ctx context.Context, sql string, args ...any) (dbiface.Rows, error) {
	return m.QueryFunc(ctx, sql, args...)
}

func (m *mockDatabaseConn) Close() {
	m.CloseFunc()
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

	for i := range dest {
		v := m.result[i]
		switch d := dest[i].(type) {
		case *string:
			if s, ok := v.(string); ok {
				*d = s
			}
		}
	}
	return nil
}

type mockRows struct {
	data       [][]any
	index      int
	err        error
	scanErrPos int // Position in data where Scan should fail
	closed     bool
}

func (m *mockRows) Next() bool {
	if m.closed {
		return false
	}
	m.index++
	return m.index <= len(m.data)
}

func (m *mockRows) Scan(dest ...any) error {
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
			}
		case *uint64:
			switch x := v.(type) {
			case uint64:
				*d = x
			}
		case *time.Time:
			if tt, ok := v.(time.Time); ok {
				*d = tt
			}
		case **time.Time: // nullable
			switch x := v.(type) {
			case *time.Time:
				*d = x
			}
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

var _ NanoID = (*mockNanoID)(nil)

type mockNanoID struct {
	GenerateFunc func(n int) (string, error)
}

func (m *mockNanoID) Generate(n int) (string, error) {
	return m.GenerateFunc(n)
}
