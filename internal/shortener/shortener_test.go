package shortener

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/anewball/urlshortener/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidURL(t *testing.T) {
	testCases := []struct {
		name     string
		url      string
		expected error
	}{
		{
			name:     "valid URL",
			url:      "http://example.com",
			expected: nil,
		},
		{
			name:     "empty URL",
			url:      "",
			expected: ErrEmptyURL,
		},
		{
			name:     "too long URL",
			url:      strings.Repeat("a", 2049),
			expected: ErrTooLong,
		},
		{
			name:     "invalid URL",
			url:      ":///invalid-url.com",
			expected: ErrParse,
		},
		{
			name:     "no scheme",
			url:      "example.com",
			expected: ErrEmptyScheme,
		},
		{
			name:     "no host",
			url:      "http://",
			expected: ErrEmptyHost,
		},
		{
			name:     "invalid scheme",
			url:      "ftp://example.com",
			expected: ErrScheme,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := isValidURL(tc.url)

			assert.ErrorIs(t, err, tc.expected)
		})
	}
}

func TestAdd(t *testing.T) {
	testCases := []struct {
		name         string
		url          string
		codeGenMock  NanoID
		conn         db.Conn
		expectedErr  error
		expectedCode string
	}{
		{
			name:         "success",
			url:          "http://example.com",
			expectedErr:  nil,
			expectedCode: "abc123",
			codeGenMock: &mockNanoID{
				GenerateFunc: func(n int) (string, error) {
					return "abc123", nil
				},
			},
			conn: &mockDatabaseConn{
				ExecFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
					return pgconn.NewCommandTag("INSERT 1"), nil
				},
			},
		},
		{
			name:         "empty URL",
			url:          "",
			expectedErr:  ErrIsValidURL,
			expectedCode: "",
			codeGenMock:  nil,
			conn:         &mockDatabaseConn{},
		},
		{
			name:         "codeGen error",
			url:          "http://example.com",
			expectedErr:  ErrGenerate,
			expectedCode: "",
			codeGenMock: &mockNanoID{
				GenerateFunc: func(n int) (string, error) {
					return "", fmt.Errorf("codeGen error")
				},
			},
			conn: &mockDatabaseConn{},
		},
		{
			name:         "exec failure",
			url:          "http://example.com",
			expectedErr:  ErrExec,
			expectedCode: "",
			codeGenMock: &mockNanoID{
				GenerateFunc: func(n int) (string, error) {
					return "abc123", nil
				},
			},
			conn: &mockDatabaseConn{
				ExecFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
					return pgconn.NewCommandTag(""), errors.New("database error")
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := New(tc.conn, tc.codeGenMock)

			shortCode, err := service.Add(context.Background(), tc.url)

			require.Equal(t, tc.expectedCode, shortCode)
			assert.ErrorIs(t, err, tc.expectedErr)
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

			service, _ := New(m, nil)
			url, err := service.Get(context.Background(), tc.shortCode)

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

			service, _ := New(m, &mockNanoID{})
			actualData, err := service.List(context.Background(), tc.limit, tc.offset)

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

			service, _ := New(m, &mockNanoID{})
			_, err := service.Delete(context.Background(), tc.shortCode)

			if tc.isError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
