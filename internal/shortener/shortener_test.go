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
		name        string
		url         string
		expectedErr error
	}{
		{
			name:        "valid URL",
			url:         "http://example.com",
			expectedErr: nil,
		},
		{
			name:        "empty URL",
			url:         "",
			expectedErr: ErrEmptyURL,
		},
		{
			name:        "too long URL",
			url:         strings.Repeat("a", 2049),
			expectedErr: ErrTooLong,
		},
		{
			name:        "invalid URL",
			url:         ":///invalid-url.com",
			expectedErr: ErrParse,
		},
		{
			name:        "no scheme",
			url:         "example.com",
			expectedErr: ErrEmptyScheme,
		},
		{
			name:        "no host",
			url:         "http://",
			expectedErr: ErrEmptyHost,
		},
		{
			name:        "invalid scheme",
			url:         "ftp://example.com",
			expectedErr: ErrScheme,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := isValidURL(tc.url)

			assert.ErrorIs(t, actualErr, tc.expectedErr)
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
			codeGenMock:  &mockNanoID{},
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

			actualShortCode, err := service.Add(context.Background(), tc.url)

			require.Equal(t, tc.expectedCode, actualShortCode)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestGet(t *testing.T) {
	testCases := []struct {
		name           string
		shortCode      string
		expectedRawURL string
		expectedErr    error
		codeGenMock    NanoID
		conn           db.Conn
	}{
		{
			name:           "success",
			shortCode:      "xK9fA3T8bfqHXEIhYkoU0M",
			expectedErr:    nil,
			expectedRawURL: "http://example.com",
			codeGenMock:    &mockNanoID{},
			conn: &mockDatabaseConn{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
					return &mockRow{result: []any{"http://example.com"}}
				},
			},
		},
		{
			name:        "empty short code",
			shortCode:   "",
			expectedErr: ErrEmptyCode,
			codeGenMock: &mockNanoID{},
			conn: &mockDatabaseConn{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
					return &mockRow{err: fmt.Errorf("short URL cannot be empty")}
				},
			},
		},
		{
			name:        "not found",
			shortCode:   "nonexistent",
			expectedErr: ErrNotFound,
			codeGenMock: &mockNanoID{},
			conn: &mockDatabaseConn{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
					return &mockRow{err: pgx.ErrNoRows}
				},
			},
		},
		{
			name:        "err tx closed",
			shortCode:   "nonexistent",
			expectedErr: ErrQuery,
			codeGenMock: &mockNanoID{},
			conn: &mockDatabaseConn{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) pgx.Row {
					return &mockRow{err: pgx.ErrTxClosed}
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := New(tc.conn, tc.codeGenMock)
			actualRawURL, err := service.Get(context.Background(), tc.shortCode)

			require.Equal(t, tc.expectedRawURL, actualRawURL)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestList(t *testing.T) {
	testCases := []struct {
		name         string
		limit        int
		offset       int
		expectedErr  error
		expectedData []URLItem
		codeGenMock  NanoID
		conn         db.Conn
	}{
		{
			name:        "success",
			limit:       10,
			offset:      0,
			expectedErr: nil,
			expectedData: []URLItem{
				{uint64(1), "http://example.com/1", "GL9VeCa", time.Date(2025, 8, 20, 12, 0, 0, 0, time.UTC), (*time.Time)(nil)},
				{uint64(2), "http://example.com/2", "GL9VeCb", time.Date(2025, 8, 20, 12, 5, 0, 0, time.UTC), (*time.Time)(nil)},
			},
			codeGenMock: &mockNanoID{},
			conn: &mockDatabaseConn{
				QueryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
					return &mockRows{
						data: [][]any{
							{uint64(1), "http://example.com/1", "GL9VeCa", time.Date(2025, 8, 20, 12, 0, 0, 0, time.UTC), (*time.Time)(nil)},
							{uint64(2), "http://example.com/2", "GL9VeCb", time.Date(2025, 8, 20, 12, 5, 0, 0, time.UTC), (*time.Time)(nil)},
						},
						index: 0,
					}, nil
				},
			},
		},
		{
			name:        "no URLs found",
			limit:       10,
			offset:      0,
			expectedErr: ErrQuery,
			codeGenMock: &mockNanoID{},
			conn: &mockDatabaseConn{
				QueryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
					return &mockRows{
						data:   [][]any{},
						index:  0,
						err:    pgx.ErrNoRows,
						closed: true,
					}, nil
				},
			},
		},
		{
			name:        "query error",
			limit:       10,
			offset:      0,
			expectedErr: ErrQuery,
			codeGenMock: &mockNanoID{},
			conn: &mockDatabaseConn{
				QueryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
					return nil, fmt.Errorf("query error")
				},
			},
		},
		{
			name:        "scan error",
			limit:       10,
			offset:      0,
			expectedErr: ErrScan,
			codeGenMock: &mockNanoID{},
			conn: &mockDatabaseConn{
				QueryFunc: func(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
					return &mockRows{
						data: [][]any{
							{"http://example.com/1"},
							{"http://example.com/2"},
						},
						scanErrPos: 1, // Simulate scan error on second row
					}, nil
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := New(tc.conn, tc.codeGenMock)
			actualData, err := service.List(context.Background(), tc.limit, tc.offset)

			require.Equal(t, tc.expectedData, actualData)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestDelete(t *testing.T) {
	testCases := []struct {
		name            string
		shortCode       string
		expectedErr     error
		expectedDeleted bool
		gen             NanoID
		conn            db.Conn
	}{
		{
			name:            "success",
			shortCode:       "xK9fA3T8bfqHXEIhYkoU0M",
			expectedDeleted: true,
			expectedErr:     nil,
			gen:             &mockNanoID{},
			conn: &mockDatabaseConn{
				ExecFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
					return pgconn.NewCommandTag("DELETE 1"), nil
				},
			},
		},
		{
			name:            "short code empty",
			shortCode:       "",
			expectedDeleted: false,
			expectedErr:     ErrEmptyCode,
			gen:             &mockNanoID{},
			conn: &mockDatabaseConn{
				ExecFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
					return pgconn.CommandTag{}, nil
				},
			},
		},
		{
			name:            "not found",
			shortCode:       "nonexistent",
			expectedDeleted: false,
			expectedErr:     ErrExec,
			gen:             &mockNanoID{},
			conn: &mockDatabaseConn{
				ExecFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
					return pgconn.CommandTag{}, pgx.ErrNoRows
				},
			},
		},
		{
			name:            "zero rows affected",
			shortCode:       "xK9fA3T8bfqHXEIhYkoU0M",
			expectedDeleted: false,
			expectedErr:     ErrNotFound,
			gen:             &mockNanoID{},
			conn: &mockDatabaseConn{
				ExecFunc: func(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error) {
					return pgconn.NewCommandTag("DELETE 0"), nil
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := New(tc.conn, tc.gen)
			actualDeleted, err := service.Delete(context.Background(), tc.shortCode)

			require.Equal(t, tc.expectedDeleted, actualDeleted)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestNew_ReturnsError_WhenDBIsNil(t *testing.T) {
	t.Parallel()

	svc, err := New(nil, nil)

	require.Nil(t, svc)
	require.Error(t, err)
}
