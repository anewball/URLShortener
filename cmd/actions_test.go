package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddActions(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		buf            bytes.Buffer
		isError        bool
		expectedError  error
		expectedResult Result
		action         Actions
		svc            shortener.URLShortener
	}{
		{
			name:           "success",
			args:           []string{"https://example.com"},
			action:         NewActions(20),
			buf:            bytes.Buffer{},
			expectedResult: Result{ShortCode: "Hpa3t2B", RawURL: "https://example.com"},
			isError:        false,
			expectedError:  nil,
			svc: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return "Hpa3t2B", nil
				},
			},
		},
		{
			name:          "error produced",
			args:          []string{"https://example.com"},
			action:        NewActions(20),
			buf:           bytes.Buffer{},
			isError:       true,
			expectedError: ErrAdd,
			svc: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return "", errors.New("something went wrong")
				},
			},
		},
		{
			name:          "error empty args",
			args:          []string{},
			action:        NewActions(20),
			buf:           bytes.Buffer{},
			isError:       true,
			expectedError: ErrLenZero,
			svc: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return "", errors.New("requires at least 1 arg(s), only received 0")
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
				require.ErrorIs(t, err, tc.expectedError)
				return
			}

			assert.NoError(t, err)
			got := tc.buf.String()
			assert.NotEmpty(t, got)

			var actualResult Result
			err = json.NewDecoder(&tc.buf).Decode(&actualResult)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedResult, actualResult)

			assert.True(t, strings.HasSuffix(got, "\n"))
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
	testCases := []struct {
		name           string
		args           []string
		buf            bytes.Buffer
		isError        bool
		expectedError  error
		expectedResult Result
		action         Actions
		svc            shortener.URLShortener
	}{
		{
			name:           "success",
			args:           []string{"Hpa3t2B"},
			action:         NewActions(20),
			buf:            bytes.Buffer{},
			expectedResult: Result{ShortCode: "Hpa3t2B", RawURL: "https://example.com"},
			isError:        false,
			expectedError:  nil,
			svc: &mockedShortener{
				getFunc: func(ctx context.Context, shortCode string) (string, error) {
					return "https://example.com", nil
				},
			},
		},
		{
			name:          "error produced",
			args:          []string{"Hpa3t2B"},
			action:        NewActions(20),
			buf:           bytes.Buffer{},
			isError:       true,
			expectedError: ErrGet,
			svc: &mockedShortener{
				getFunc: func(ctx context.Context, url string) (string, error) {
					return "", errors.New("something went wrong")
				},
			},
		},
		{
			name:          "error empty args",
			args:          []string{},
			action:        NewActions(20),
			buf:           bytes.Buffer{},
			isError:       true,
			expectedError: ErrLenZero,
			svc: &mockedShortener{
				getFunc: func(ctx context.Context, url string) (string, error) {
					return "", errors.New("requires at least 1 arg(s), only received 0")
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
				require.ErrorIs(t, err, tc.expectedError)
				return
			}

			assert.NoError(t, err)
			got := tc.buf.String()
			assert.NotEmpty(t, got)

			var actualResult Result
			err = json.NewDecoder(&tc.buf).Decode(&actualResult)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedResult, actualResult)

			assert.True(t, strings.HasSuffix(got, "\n"))
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
	testCases := []struct {
		name           string
		args           []string
		buf            bytes.Buffer
		isError        bool
		expectedError  error
		expectedResult DeleteResponse
		action         Actions
		svc            shortener.URLShortener
	}{
		{
			name:           "success",
			args:           []string{"Hpa3t2B"},
			action:         NewActions(20),
			buf:            bytes.Buffer{},
			expectedResult: DeleteResponse{Deleted: true, ShortCode: "Hpa3t2B"},
			isError:        false,
			expectedError:  nil,
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, shortCode string) (bool, error) {
					return true, nil
				},
			},
		},
		{
			name:          "error produced",
			args:          []string{"Hpa3t2B"},
			action:        NewActions(20),
			buf:           bytes.Buffer{},
			isError:       true,
			expectedError: ErrDelete,
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, errors.New("something went wrong")
				},
			},
		},
		{
			name:          "error empty args",
			args:          []string{},
			action:        NewActions(20),
			buf:           bytes.Buffer{},
			isError:       true,
			expectedError: ErrLenZero,
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, errors.New("requires at least 1 arg(s), only received 0")
				},
			},
		},
		{
			name:          "when row does not exists",
			args:          []string{"Hpa3t2B"},
			action:        NewActions(20),
			buf:           bytes.Buffer{},
			isError:       true,
			expectedError: ErrNotFound,
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
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			got := tc.buf.String()
			assert.NotEmpty(t, got)

			var actualResult DeleteResponse
			err = json.NewDecoder(&tc.buf).Decode(&actualResult)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedResult, actualResult)

			assert.True(t, strings.HasSuffix(got, "\n"))
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
	testCases := []struct {
		name           string
		offset         int
		limit          int
		buf            bytes.Buffer
		isError        bool
		expectedError  error
		expectedResult []Result
		action         Actions
		svc            shortener.URLShortener
	}{
		{
			name:   "success",
			offset: 0,
			limit:  2,
			action: NewActions(20),
			buf:    bytes.Buffer{},
			expectedResult: []Result{
				{RawURL: "https://anewball.com", ShortCode: "nMHdgTh"},
				{RawURL: "https://jayden.newball.com", ShortCode: "k5aBWD5"},
			},
			isError:       false,
			expectedError: nil,
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
			expectedResult: []Result{
				{RawURL: "https://anewball.com", ShortCode: "nMHdgTh"},
				{RawURL: "https://jayden.newball.com", ShortCode: "k5aBWD5"},
			},
			isError:       false,
			expectedError: nil,
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
			name:           "limit less than zero",
			offset:         0,
			limit:          -2,
			action:         NewActions(20),
			buf:            bytes.Buffer{},
			expectedResult: []Result{},
			isError:        true,
			expectedError:  ErrLimit,
			svc:            &mockedShortener{},
		},
		{
			name:           "offset less than zero",
			offset:         -2,
			limit:          2,
			action:         NewActions(20),
			buf:            bytes.Buffer{},
			expectedResult: []Result{},
			isError:        true,
			expectedError:  ErrNegativeOffset,
			svc:            &mockedShortener{},
		},
		{
			name:           "list returned error",
			offset:         0,
			limit:          2,
			action:         NewActions(20),
			buf:            bytes.Buffer{},
			expectedResult: []Result{},
			isError:        true,
			expectedError:  ErrList,
			svc: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{}, errors.New("could not retrieve data")
				},
			},
		},
		{
			name:           "when limit exceeds max",
			offset:         -2,
			limit:          21,
			action:         NewActions(20),
			buf:            bytes.Buffer{},
			expectedResult: []Result{},
			isError:        true,
			expectedError:  ErrLimit,
			svc:            &mockedShortener{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := tc.action.ListAction(ctx, tc.limit, tc.offset, &tc.buf, tc.svc)

			if tc.isError {
				require.ErrorIs(t, err, tc.expectedError)
				return
			}

			assert.NoError(t, err)
			got := tc.buf.String()
			assert.NotEmpty(t, got)

			var actualResult []Result
			err = json.NewDecoder(&tc.buf).Decode(&actualResult)
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedResult, actualResult)

			assert.True(t, strings.HasSuffix(got, "\n"))
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
