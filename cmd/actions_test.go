package cmd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/anewball/urlshortener/internal/jsonutil"
	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddActions(t *testing.T) {
	listMaxLimit := 20
	shortCode := "Hpa3t2B"
	testCases := []struct {
		name                   string
		args                   []string
		buf                    bytes.Buffer
		isError                bool
		expectedErrorResponse  ErrorResponse
		expectedResultResponse ResultResponse
		action                 Actions
		svc                    shortener.URLShortener
	}{
		{
			name:                   "success",
			args:                   []string{"https://example.com"},
			action:                 NewActions(listMaxLimit),
			buf:                    bytes.Buffer{},
			expectedResultResponse: ResultResponse{ShortCode: shortCode, RawURL: "https://example.com"},
			isError:                false,
			expectedErrorResponse:  ErrorResponse{},
			svc: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return shortCode, nil
				},
			},
		},
		{
			name:                  "zero args",
			args:                  []string{},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrLenZero.Error()},
			svc:                   &mockedShortener{},
		},
		{
			name:                  "invalid url",
			args:                  []string{"https://example.com"},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrURL.Error(), Details: shortener.ErrIsValidURL.Error()},
			svc: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return "", shortener.ErrIsValidURL
				},
			},
		},
		{
			name:                  "error empty args",
			args:                  []string{"https://example.com"},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrGenCode.Error(), Details: shortener.ErrGenerate.Error()},
			svc: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return "", shortener.ErrGenerate
				},
			},
		},
		{
			name:                  "could not add url",
			args:                  []string{"https://example.com"},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrAdd.Error(), Details: shortener.ErrQueryRow.Error()},
			svc: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return "", shortener.ErrQueryRow
				},
			},
		},
		{
			name:                  "error not supported",
			args:                  []string{"https://example.com"},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrUnsupported.Error(), Details: "Failed to add URL"},
			svc: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return "", errors.New("Failed to add URL")
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := tc.action.AddAction(ctx, &tc.buf, tc.svc, tc.args)

			if tc.isError {
				var actualErrorResponse ErrorResponse
				jsonutil.ReadJSON(&tc.buf, &actualErrorResponse)
				assert.Equal(t, tc.expectedErrorResponse, actualErrorResponse)
				return
			}

			assert.NoError(t, err)

			var actualResultResponse ResultResponse
			jsonutil.ReadJSON(&tc.buf, &actualResultResponse)

			assert.Equal(t, tc.expectedResultResponse, actualResultResponse)
		})
	}
}

func TestNewAdd(t *testing.T) {
	called := false
	var gotCtx context.Context
	var gotOut io.Writer
	var gotArgs []string

	mActions := &mockedActions{
		AddActionFunc: func(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error {
			called = true
			gotCtx = ctx
			gotOut = out
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	cmd := NewAdd(mActions, &mockedShortener{})

	assert.Equal(t, "add <url>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	args := []string{"https://example.com"}

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)

	// Execute the command exactly like a user would
	require.NoError(t, cmd.ExecuteContext(context.Background()))

	// Assertions on wiring
	assert.True(t, called, "AddAction should be invoked")
	assert.Equal(t, args, gotArgs)
	assert.Same(t, buf, gotOut)
	assert.NotNil(t, gotCtx)
}

func TestGetAction(t *testing.T) {
	listMaxLimit := 20
	shortCode := "Hpa3t2B"
	testCases := []struct {
		name                   string
		args                   []string
		buf                    bytes.Buffer
		isError                bool
		expectedErrorResponse  ErrorResponse
		expectedResultResponse ResultResponse
		action                 Actions
		svc                    shortener.URLShortener
	}{
		{
			name:                   "success",
			args:                   []string{shortCode},
			action:                 NewActions(listMaxLimit),
			buf:                    bytes.Buffer{},
			expectedResultResponse: ResultResponse{ShortCode: shortCode, RawURL: "https://example.com"},
			isError:                false,
			expectedErrorResponse:  ErrorResponse{},
			svc: &mockedShortener{
				getFunc: func(ctx context.Context, shortCode string) (string, error) {
					return "https://example.com", nil
				},
			},
		},
		{
			name:                  "zero args",
			args:                  []string{},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrLenZero.Error()},
			svc:                   &mockedShortener{},
		},
		{
			name:                  "error empty short code",
			args:                  []string{""},
			action:                NewActions(20),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrShorCode.Error(), Details: shortener.ErrEmptyShortCode.Error()},
			svc: &mockedShortener{
				getFunc: func(ctx context.Context, url string) (string, error) {
					return "", shortener.ErrEmptyShortCode
				},
			},
		},
		{
			name:                  "error not found",
			args:                  []string{shortCode},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Sprintf("%s: %s", ErrNotFound, shortCode), Details: shortener.ErrNotFound.Error()},
			svc: &mockedShortener{
				getFunc: func(ctx context.Context, url string) (string, error) {
					return "", shortener.ErrNotFound
				},
			},
		},
		{
			name:                  "error query",
			args:                  []string{shortCode},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrTimeout.Error(), Details: shortener.ErrQuery.Error()},
			svc: &mockedShortener{
				getFunc: func(ctx context.Context, url string) (string, error) {
					return "", shortener.ErrQuery
				},
			},
		},
		{
			name:                  "error",
			args:                  []string{shortCode},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrUnexpected.Error(), Details: "Something went wrong"},
			svc: &mockedShortener{
				getFunc: func(ctx context.Context, url string) (string, error) {
					return "", errors.New("Something went wrong")
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := tc.action.GetAction(ctx, &tc.buf, tc.svc, tc.args)

			if tc.isError {
				var actualErrorResponse ErrorResponse
				jsonutil.ReadJSON(&tc.buf, &actualErrorResponse)
				assert.Equal(t, tc.expectedErrorResponse, actualErrorResponse)
				return
			}

			assert.NoError(t, err)

			var actualResultResponse ResultResponse
			jsonutil.ReadJSON(&tc.buf, &actualResultResponse)

			assert.Equal(t, tc.expectedResultResponse, actualResultResponse)
		})
	}
}

func TestNewGet(t *testing.T) {
	called := false
	var gotCtx context.Context
	var gotOut io.Writer
	var gotArgs []string

	mActions := &mockedActions{
		GetActionFunc: func(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error {
			called = true
			gotCtx = ctx
			gotOut = out
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	cmd := NewGet(mActions, &mockedShortener{})

	assert.Equal(t, "get <code>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	args := []string{"Hpa3t2B"}

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)

	// Execute the command exactly like a user would
	require.NoError(t, cmd.ExecuteContext(context.Background()))

	// Assertions on wiring
	assert.True(t, called, "GetAction should be invoked")
	assert.Equal(t, args, gotArgs)
	assert.Same(t, buf, gotOut)
	assert.NotNil(t, gotCtx)
}

func TestDeleteAction(t *testing.T) {
	listMaxLimit := 20
	shortCode := "Hpa3t2B"
	testCases := []struct {
		name                   string
		args                   []string
		buf                    bytes.Buffer
		isError                bool
		expectedErrorResponse  ErrorResponse
		expectedDeleteResponse DeleteResponse
		action                 Actions
		svc                    shortener.URLShortener
	}{
		{
			name:                   "success",
			args:                   []string{shortCode},
			action:                 NewActions(listMaxLimit),
			buf:                    bytes.Buffer{},
			expectedDeleteResponse: DeleteResponse{Deleted: true, ShortCode: shortCode},
			isError:                false,
			expectedErrorResponse:  ErrorResponse{},
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, shortCode string) (bool, error) {
					return true, nil
				},
			},
		},
		{
			name:                  "zero args",
			args:                  []string{},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrLenZero.Error()},
			svc:                   &mockedShortener{},
		},
		{
			name:                  "error empty short code",
			args:                  []string{""},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: "A short code is required"},
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, shortener.ErrEmptyShortCode
				},
			},
		},
		{
			name:                  "error with exec",
			args:                  []string{shortCode},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Sprintf("A problem occurs when deleting short code: %s", shortCode)},
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, shortener.ErrExec
				},
			},
		},
		{
			name:                  "error URL not found",
			args:                  []string{shortCode},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Sprintf("Could not delete URL with short code %s", shortCode)},
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, shortener.ErrNotFound
				},
			},
		},
		{
			name:                  "unknown error",
			args:                  []string{shortCode},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Sprintf("Service could not delete URL with short code %s", shortCode)},
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, errors.New("unknown error")
				},
			},
		},
		{
			name:                  "when deleted variable is false",
			args:                  []string{shortCode},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Sprintf("Problem deleting URL with short code %q", shortCode)},
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, nil
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := tc.action.DeleteAction(ctx, &tc.buf, tc.svc, tc.args)

			if tc.isError {
				var actualErrorResponse ErrorResponse
				jsonutil.ReadJSON(&tc.buf, &actualErrorResponse)
				assert.Equal(t, tc.expectedErrorResponse, actualErrorResponse)
				return
			}

			assert.NoError(t, err)

			var actualDeleteResponse DeleteResponse
			jsonutil.ReadJSON(&tc.buf, &actualDeleteResponse)

			assert.Equal(t, tc.expectedDeleteResponse, actualDeleteResponse)
		})
	}
}

func TestNewDelete(t *testing.T) {
	called := false
	var gotCtx context.Context
	var gotOut io.Writer
	var gotArgs []string

	mActions := &mockedActions{
		DeleteActionFunc: func(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error {
			called = true
			gotCtx = ctx
			gotOut = out
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	cmd := NewDelete(mActions, &mockedShortener{})

	assert.Equal(t, "delete <code>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	args := []string{"Hpa3t2B"}

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)

	// Execute the command exactly like a user would
	require.NoError(t, cmd.ExecuteContext(context.Background()))

	// Assertions on wiring
	assert.True(t, called, "DeleteAction should be invoked")
	assert.Equal(t, args, gotArgs)
	assert.Same(t, buf, gotOut)
	assert.NotNil(t, gotCtx)
}

func TestListAction(t *testing.T) {
	listMaxLimit := 20
	testCases := []struct {
		name                  string
		offset                int
		limit                 int
		buf                   bytes.Buffer
		isError               bool
		expectedErrorResponse ErrorResponse
		expectedListResponse  ListResponse
		action                Actions
		svc                   shortener.URLShortener
	}{
		{
			name:   "success",
			offset: 0,
			limit:  2,
			action: NewActions(listMaxLimit),
			buf:    bytes.Buffer{},
			expectedListResponse: ListResponse{
				Items: []ResultResponse{
					{RawURL: "https://anewball.com", ShortCode: "nMHdgTh"},
					{RawURL: "https://jayden.newball.com", ShortCode: "k5aBWD5"},
				}, Count: 2, Limit: 2, Offset: 0,
			},
			isError:               false,
			expectedErrorResponse: ErrorResponse{},
			svc: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{
						{ID: 1, OriginalURL: "https://anewball.com", ShortCode: "nMHdgTh", CreatedAt: time.Date(2025, time.August, 25, 14, 30, 0, 0, time.UTC), ExpiresAt: nil},
						{ID: 2, OriginalURL: "https://jayden.newball.com", ShortCode: "k5aBWD5", CreatedAt: time.Date(2025, time.August, 25, 14, 3, 0, 0, time.UTC), ExpiresAt: nil},
					}, nil
				},
			},
		},
		{
			name:   "success when limit max is zero",
			offset: 0,
			limit:  2,
			action: NewActions(0),
			buf:    bytes.Buffer{},
			expectedListResponse: ListResponse{
				Items: []ResultResponse{
					{RawURL: "https://anewball.com", ShortCode: "nMHdgTh"},
					{RawURL: "https://jayden.newball.com", ShortCode: "k5aBWD5"},
				}, Count: 2, Limit: 2, Offset: 0,
			},
			isError:               false,
			expectedErrorResponse: ErrorResponse{},
			svc: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{
						{ID: 1, OriginalURL: "https://anewball.com", ShortCode: "nMHdgTh", CreatedAt: time.Date(2025, time.August, 25, 14, 30, 0, 0, time.UTC), ExpiresAt: nil},
						{ID: 2, OriginalURL: "https://jayden.newball.com", ShortCode: "k5aBWD5", CreatedAt: time.Date(2025, time.August, 25, 14, 3, 0, 0, time.UTC), ExpiresAt: nil},
					}, nil
				},
			},
		},
		{
			name:                  "limit less than zero",
			offset:                0,
			limit:                 -2,
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			expectedListResponse:  ListResponse{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Sprintf("%s: %d", ErrLimit.Error(), listMaxLimit)},
			svc:                   &mockedShortener{},
		},
		{
			name:                  "offset less than zero",
			offset:                -2,
			limit:                 2,
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			expectedListResponse:  ListResponse{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Errorf("%w: %d", ErrNegativeOffset, -2).Error()},
			svc:                   &mockedShortener{},
		},
		{
			name:                  "error query",
			offset:                0,
			limit:                 2,
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			expectedListResponse:  ListResponse{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Sprintf("Failed to retrieve URLs with limit: %d and offset: %d", 2, 0)},
			svc: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{}, shortener.ErrQuery
				},
			},
		},
		{
			name:                  "error scan",
			offset:                0,
			limit:                 2,
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			expectedListResponse:  ListResponse{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Sprintf("Failed to smarshal URLs with limit: %d and offset: %d", 2, 0)},
			svc: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{}, shortener.ErrScan
				},
			},
		},
		{
			name:                  "error rows",
			offset:                0,
			limit:                 2,
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			expectedListResponse:  ListResponse{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Sprintf("An error occurs when smarshal URLs with limit: %d and offset: %d", 2, 0)},
			svc: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{}, shortener.ErrRows
				},
			},
		},
		{
			name:                  "error rows",
			offset:                0,
			limit:                 2,
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			expectedListResponse:  ListResponse{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Sprintf("An error occurs when retrieving URLs from limit: %d and offset: %d", 2, 0)},
			svc: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{}, errors.New("something went wrong")
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := tc.action.ListAction(ctx, tc.limit, tc.offset, &tc.buf, tc.svc)

			if tc.isError {
				var actualErrorResponse ErrorResponse
				jsonutil.ReadJSON(&tc.buf, &actualErrorResponse)
				assert.Equal(t, tc.expectedErrorResponse, actualErrorResponse)
				return
			}

			assert.NoError(t, err)

			var actualListResponse ListResponse
			jsonutil.ReadJSON(&tc.buf, &actualListResponse)

			assert.Equal(t, tc.expectedListResponse, actualListResponse)
		})
	}
}

func TestNewList(t *testing.T) {
	called := false
	var gotCtx context.Context
	var gotOut io.Writer

	mActions := &mockedActions{
		ListActionFunc: func(ctx context.Context, limit int, offset int, out io.Writer, svc shortener.URLShortener) error {
			called = true
			gotCtx = ctx
			gotOut = out
			return nil
		},
	}

	cmd := NewList(mActions, &mockedShortener{})

	assert.Equal(t, "list", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--offset", "0", "--limit", "2"})

	// Execute the command exactly like a user would
	require.NoError(t, cmd.Execute())

	// Assertions on wiring
	assert.True(t, called, "ListAction should be invoked")
	assert.Same(t, buf, gotOut)
	assert.NotNil(t, gotCtx)
}

func TestNewRoot(t *testing.T) {
	cmd := NewRoot(&mockedActions{}, &mockedShortener{})

	assert.Equal(t, "urlshortener", cmd.Use)
}
