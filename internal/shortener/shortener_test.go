package shortener

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func TestAdd(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		isError  bool
		execFunc func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	}{
		{
			name:    "success",
			url:     "http://example.com",
			isError: false,
			execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
				return pgconn.NewCommandTag("INSERT 1"), nil
			},
		},
		{
			name:    "empty URL",
			url:     "",
			isError: true,
			execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, fmt.Errorf("URL cannot be empty")
			},
		},
		{
			name:    "failure",
			url:     "http://example.com",
			isError: true,
			execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
				return pgconn.NewCommandTag(""), errors.New("database error")
			},
		},
		{
			name:    "error 23505",
			url:     "http://example.com",
			isError: true,
			execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, &pgconn.PgError{
					Code:    "23505",
					Message: "duplicate key value violates unique constraint",
					Detail:  "Key (short_code)=(abc123) already exists.",
				}
			},
		},
		{
			name:    "No URL scheme",
			url:     "example.com",
			isError: true,
			execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, errors.New("No URL scheme")
			},
		},
				{
			name:    "Invalid Scheme",
			url:     "ppp://example.com",
			isError: true,
			execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, errors.New("Invalid URL scheme")
			},
		},
		{
			name:    "No Scheme with invalid characters",
			url:     ":/invalid-url",
			isError: true,
			execFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, errors.New("invalid URL format")
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockDatabaseConn{ExecFunc: tc.execFunc}

			s := New(m)
			shortCode, err := s.Add(context.Background(), tc.url)

			if tc.isError {
				require.Error(t, err)
				require.Empty(t, shortCode)
			} else {
				require.NoError(t, err)
				require.NotEmpty(t, shortCode)
			}
		})
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		name         string
		shortCode    string
		isError      bool
		url          string
		queryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
	}{
		{
			name:      "success",
			shortCode: "xK9fA3T8bfqHXEIhYkoU0M",
			isError:   false,
			url:       "http://example.com",
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{result: []any{"http://example.com"}}
			},
		},
		{
			name:      "empty short code",
			shortCode: "",
			isError:   true,
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{err: fmt.Errorf("short URL cannot be empty")}
			},
		},
		{
			name:      "not found",
			shortCode: "nonexistent",
			isError:   true,
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{err: pgx.ErrNoRows}
			},
		},
		{
			name:      "err tx closed",
			shortCode: "nonexistent",
			isError:   true,
			queryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
				return &mockRow{err: pgx.ErrTxClosed}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockDatabaseConn{QueryRowFunc: tc.queryRowFunc}

			s := New(m)
			url, err := s.Get(context.Background(), tc.shortCode)

			if tc.isError {
				require.Error(t, err)
				require.Empty(t, url)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.url, url)
			}
		})
	}
}

func TestList(t *testing.T) {
	testCases := []struct {
		name         string
		limit        int
		offset       int
		isError      bool
		expectedData []URLItem
		queryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	}{
		{
			name:    "success",
			limit:   10,
			offset:  0,
			isError: false,
			expectedData: []URLItem{
				{uint64(1), "http://example.com/1", "GL9VeCa", time.Date(2025, 8, 20, 12, 0, 0, 0, time.UTC), (*time.Time)(nil)},
				{uint64(2), "http://example.com/2", "GL9VeCb", time.Date(2025, 8, 20, 12, 5, 0, 0, time.UTC), (*time.Time)(nil)},
			},
			queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
				return &mockRows{
					data: [][]any{
						{uint64(1), "http://example.com/1", "GL9VeCa", time.Date(2025, 8, 20, 12, 0, 0, 0, time.UTC), (*time.Time)(nil)},
						{uint64(2), "http://example.com/2", "GL9VeCb", time.Date(2025, 8, 20, 12, 5, 0, 0, time.UTC), (*time.Time)(nil)},
					},
					index: 0,
				}, nil
			},
		},
		{
			name:    "no URLs found",
			limit:   10,
			offset:  0,
			isError: true,
			queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
				return &mockRows{
					data:   [][]any{},
					index:  0,
					err:    pgx.ErrNoRows,
					closed: true,
				}, nil
			},
		},
		{
			name:    "query error",
			limit:   10,
			offset:  0,
			isError: true,
			queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
				return nil, fmt.Errorf("query error")
			},
		},
		{
			name:    "scan error",
			limit:   10,
			offset:  0,
			isError: true,
			queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
				return &mockRows{
					data: [][]any{
						{"http://example.com/1"},
						{"http://example.com/2"},
					},
					scanErrPos: 1, // Simulate scan error on second row
				}, nil
			},
		},
		{
			name:    "URLs empty",
			limit:   10,
			offset:  0,
			isError: true,
			queryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
				return &mockRows{
					index:  0,
					closed: true,
				}, nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockDatabaseConn{QueryFunc: tc.queryFunc}

			s := New(m)
			actualData, err := s.List(context.Background(), tc.limit, tc.offset)

			if tc.isError {
				require.Error(t, err)
				require.Empty(t, actualData)
			} else {
				require.NoError(t, err)
				require.Equal(t, tc.expectedData, actualData)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		name      string
		shortCode string
		isError   bool
		execFunc  func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	}{
		{
			name:      "success",
			shortCode: "xK9fA3T8bfqHXEIhYkoU0M",
			isError:   false,
			execFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
		},
		{
			name:      "short code empty",
			shortCode: "",
			isError:   true,
			execFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, nil
			},
		},
		{
			name:      "not found",
			shortCode: "nonexistent",
			isError:   true,
			execFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, pgx.ErrNoRows
			},
		},
		{
			name:      "err tx closed",
			shortCode: "nonexistent",
			isError:   true,
			execFunc: func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
				return pgconn.CommandTag{}, pgx.ErrTxClosed
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			m := &mockDatabaseConn{ExecFunc: tc.execFunc}

			s := New(m)
			_, err := s.Delete(context.Background(), tc.shortCode)

			if tc.isError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
