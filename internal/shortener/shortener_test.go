package shortener

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/anewball/urlshortener/internal/dbiface"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsValidURL(t *testing.T) {
	testCases := []struct {
		name        string
		rawURL      string
		expectedErr error
	}{
		{
			name:        "valid URL",
			rawURL:      "http://example.com",
			expectedErr: nil,
		},
		{
			name:        "empty URL",
			rawURL:      "",
			expectedErr: ErrEmptyURL,
		},
		{
			name:        "too long URL",
			rawURL:      strings.Repeat("a", 2049),
			expectedErr: ErrTooLong,
		},
		{
			name:        "invalid URL",
			rawURL:      ":///invalid-url.com",
			expectedErr: ErrParse,
		},
		{
			name:        "no scheme",
			rawURL:      "example.com",
			expectedErr: ErrEmptyScheme,
		},
		{
			name:        "no host",
			rawURL:      "http://",
			expectedErr: ErrEmptyHost,
		},
		{
			name:        "invalid scheme",
			rawURL:      "ftp://example.com",
			expectedErr: ErrScheme,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualErr := isValidURL(tc.rawURL)

			assert.ErrorIs(t, actualErr, tc.expectedErr)
		})
	}
}

func TestAdd(t *testing.T) {
	testCases := []struct {
		name              string
		rawURL            string
		gen               NanoID
		querier           dbiface.Querier
		expectedErr       error
		expectedShortCode string
	}{
		{
			name:              "success",
			rawURL:            "http://example.com",
			expectedErr:       nil,
			expectedShortCode: "abc123",
			gen: &mockNanoID{
				GenerateFunc: func(n int) (string, error) {
					return "abc123", nil
				},
			},
			querier: &mockQuerier{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) dbiface.Row {
					return &mockRow{result: []any{"abc123"}}
				},
			},
		},
		{
			name:              "empty URL",
			rawURL:            "",
			expectedErr:       ErrIsValidURL,
			expectedShortCode: "",
			gen:               &mockNanoID{},
			querier:           &mockQuerier{},
		},
		{
			name:              "codeGen error",
			rawURL:            "http://example.com",
			expectedErr:       ErrGenerate,
			expectedShortCode: "",
			gen: &mockNanoID{
				GenerateFunc: func(n int) (string, error) {
					return "", fmt.Errorf("codeGen error")
				},
			},
			querier: &mockQuerier{},
		},
		{
			name:              "exec failure",
			rawURL:            "http://example.com",
			expectedErr:       ErrExec,
			expectedShortCode: "",
			gen: &mockNanoID{
				GenerateFunc: func(n int) (string, error) {
					return "abc123", nil
				},
			},
			querier: &mockQuerier{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) dbiface.Row {
					return &mockRow{err: errors.New("database error")}
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := New(tc.querier, tc.gen)

			actualShortCode, err := service.Add(context.Background(), tc.rawURL)

			require.Equal(t, tc.expectedShortCode, actualShortCode)
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
		gen            NanoID
		querier        dbiface.Querier
	}{
		{
			name:           "success",
			shortCode:      "xK9fA3T8bfqHXEIhYkoU0M",
			expectedErr:    nil,
			expectedRawURL: "http://example.com",
			gen:            &mockNanoID{},
			querier: &mockQuerier{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) dbiface.Row {
					return &mockRow{result: []any{"http://example.com"}}
				},
			},
		},
		{
			name:        "empty short code",
			shortCode:   "",
			expectedErr: ErrEmptyCode,
			gen:         &mockNanoID{},
			querier: &mockQuerier{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) dbiface.Row {
					return &mockRow{err: fmt.Errorf("short URL cannot be empty")}
				},
			},
		},
		{
			name:        "not found",
			shortCode:   "nonexistent",
			expectedErr: ErrNotFound,
			gen:         &mockNanoID{},
			querier: &mockQuerier{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) dbiface.Row {
					return &mockRow{err: ErrNotFound}
				},
			},
		},
		{
			name:        "err tx closed",
			shortCode:   "nonexistent",
			expectedErr: ErrQuery,
			gen:         &mockNanoID{},
			querier: &mockQuerier{
				QueryRowFunc: func(ctx context.Context, sql string, args ...any) dbiface.Row {
					return &mockRow{err: ErrQuery}
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := New(tc.querier, tc.gen)
			actualRawURL, err := service.Get(context.Background(), tc.shortCode)

			require.Equal(t, tc.expectedRawURL, actualRawURL)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestList(t *testing.T) {
	testCases := []struct {
		name          string
		limit         int
		offset        int
		expectedErr   error
		expectedItems []URLItem
		gen           NanoID
		querier       dbiface.Querier
	}{
		{
			name:  "success",
			limit: 10, offset: 0,
			expectedErr: nil,
			expectedItems: []URLItem{
				{uint64(1), "http://example.com/1", "GL9VeCa", time.Date(2025, 8, 20, 12, 0, 0, 0, time.UTC), (*time.Time)(nil)},
				{uint64(2), "http://example.com/2", "GL9VeCb", time.Date(2025, 8, 20, 12, 5, 0, 0, time.UTC), (*time.Time)(nil)},
			},
			gen: &mockNanoID{},
			querier: &mockQuerier{
				QueryFunc: func(ctx context.Context, sql string, args ...any) (dbiface.Rows, error) {
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
			name:  "rows error",
			limit: 10, offset: 0,
			expectedErr: ErrQuery,
			gen:         &mockNanoID{},
			querier: &mockQuerier{
				QueryFunc: func(ctx context.Context, sql string, args ...any) (dbiface.Rows, error) {
					return &mockRows{
						data:   [][]any{},
						index:  0,
						err:    ErrQuery,
						closed: true,
					}, nil
				},
			},
		},
		{
			name:  "query error",
			limit: 10, offset: 0,
			expectedErr: ErrQuery,
			gen:         &mockNanoID{},
			querier: &mockQuerier{
				QueryFunc: func(ctx context.Context, sql string, args ...any) (dbiface.Rows, error) {
					return nil, fmt.Errorf("query error")
				},
			},
		},
		{
			name:  "scan error",
			limit: 10, offset: 0,
			expectedErr: ErrScan,
			gen:         &mockNanoID{},
			querier: &mockQuerier{
				QueryFunc: func(ctx context.Context, sql string, args ...any) (dbiface.Rows, error) {
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
		{
			name:  "empty results",
			limit: 10, offset: 0,
			expectedErr:   nil,
			gen:           &mockNanoID{},
			expectedItems: []URLItem{},
			querier: &mockQuerier{
				QueryFunc: func(ctx context.Context, sql string, args ...any) (dbiface.Rows, error) {
					return &mockRows{data: [][]any{}}, nil
				},
			},
		},
		{
			name:  "rows iteration error",
			limit: 10, offset: 0,
			expectedErr: ErrQuery,
			gen:         &mockNanoID{},
			querier: &mockQuerier{
				QueryFunc: func(ctx context.Context, sql string, args ...any) (dbiface.Rows, error) {
					return &mockRows{data: [][]any{}, err: ErrScan}, nil
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := New(tc.querier, tc.gen)
			actualItems, err := service.List(context.Background(), tc.limit, tc.offset)

			require.Equal(t, tc.expectedItems, actualItems)
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
		querier         dbiface.Querier
	}{
		{
			name:            "success",
			shortCode:       "xK9fA3T8bfqHXEIhYkoU0M",
			expectedDeleted: true,
			expectedErr:     nil,
			gen:             &mockNanoID{},
			querier: &mockQuerier{
				ExecFunc: func(ctx context.Context, sql string, arguments ...any) (dbiface.CommandResult, error) {
					return dbiface.CommandResult(&mockCommandResult{rowsAffected: 1}), nil
				},
			},
		},
		{
			name:            "short code empty",
			shortCode:       "",
			expectedDeleted: false,
			expectedErr:     ErrEmptyCode,
			gen:             &mockNanoID{},
			querier: &mockQuerier{
				ExecFunc: func(ctx context.Context, sql string, arguments ...any) (dbiface.CommandResult, error) {
					return dbiface.CommandResult(&mockCommandResult{rowsAffected: 0}), nil
				},
			},
		},
		{
			name:            "not found",
			shortCode:       "nonexistent",
			expectedDeleted: false,
			expectedErr:     ErrExec,
			gen:             &mockNanoID{},
			querier: &mockQuerier{
				ExecFunc: func(ctx context.Context, sql string, arguments ...any) (dbiface.CommandResult, error) {
					return dbiface.CommandResult(&mockCommandResult{rowsAffected: 0}), ErrNotFound
				},
			},
		},
		{
			name:            "zero rows affected",
			shortCode:       "xK9fA3T8bfqHXEIhYkoU0M",
			expectedDeleted: false,
			expectedErr:     ErrNotFound,
			gen:             &mockNanoID{},
			querier: &mockQuerier{
				ExecFunc: func(ctx context.Context, sql string, arguments ...any) (dbiface.CommandResult, error) {
					return dbiface.CommandResult(&mockCommandResult{rowsAffected: 0}), nil
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			service, _ := New(tc.querier, tc.gen)
			actualDeleted, err := service.Delete(context.Background(), tc.shortCode)

			require.Equal(t, tc.expectedDeleted, actualDeleted)
			assert.ErrorIs(t, err, tc.expectedErr)
		})
	}
}

func TestNew_ReturnsError_WhenDBIsNil(t *testing.T) {
	testCases := []struct {
		name        string
		db          dbiface.Querier
		gen         NanoID
		expectedErr error
		isErrNil    bool
	}{
		{
			name:        "success",
			db:          &mockQuerier{},
			gen:         &mockNanoID{},
			expectedErr: nil,
			isErrNil:    true,
		},
		{
			name:        "error when db is nil",
			db:          nil,
			gen:         &mockNanoID{},
			expectedErr: ErrDBNil,
			isErrNil:    false,
		},
		{
			name:        "error when gen is nil",
			db:          &mockQuerier{},
			gen:         nil,
			expectedErr: ErrNanoIDNil,
			isErrNil:    false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, actualErr := New(tc.db, tc.gen)

			if tc.isErrNil {
				require.NotNil(t, svc)
				require.NoError(t, actualErr)
			} else {
				require.Nil(t, svc)
				assert.ErrorIs(t, actualErr, tc.expectedErr)
			}
		})
	}
}

func TestClose(t *testing.T) {
	db := &mockQuerier{
		CloseFunc: func() {
			log.Println("mock db closed")
		},
	}
	require.NotNil(t, db)
	db.Close()
}
