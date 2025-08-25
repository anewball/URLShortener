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
		expectedResult Result
		actionFunction func(context.Context, io.Writer, *App, []string) error
		shor           shortener.Shortener
	}{
		{
			name:           "success",
			args:           []string{"https://example.com"},
			actionFunction: addAction,
			buf:            bytes.Buffer{},
			expectedResult: Result{Code: "Hpa3t2B", Url: "https://example.com"},
			isError:        false,
			shor: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return "Hpa3t2B", nil
				},
			},
		},
		{
			name:           "error produced",
			args:           []string{"https://example.com"},
			actionFunction: addAction,
			buf:            bytes.Buffer{},
			isError:        true,
			shor: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return "", errors.New("something went wrong")
				},
			},
		},
		{
			name:           "error empty args",
			args:           []string{},
			actionFunction: addAction,
			buf:            bytes.Buffer{},
			isError:        true,
			shor: &mockedShortener{
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

			app := &App{S: tc.shor}

			err := tc.actionFunction(ctx, &tc.buf, app, tc.args)

			if tc.isError {
				assert.Error(t, err)
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
	t.Cleanup(func() { addActionFunc = addAction })

	called := false
	var gotCtx context.Context
	var gotOut io.Writer
	var gotArgs []string

	addActionFunc = func(ctx context.Context, out io.Writer, a *App, args []string) error {
		called = true
		gotCtx = ctx
		gotOut = out
		gotArgs = append([]string(nil), args...)
		return nil
	}

	app := &App{S: &mockedShortener{}}
	cmd := NewAdd(app)

	assert.Equal(t, "add <url>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"https://example.com"})

	// Execute the command exactly like a user would
	require.NoError(t, cmd.ExecuteContext(context.Background()))

	// Assertions on wiring
	assert.True(t, called, "addActionFn should be invoked")
	assert.Equal(t, []string{"https://example.com"}, gotArgs)
	assert.Same(t, buf, gotOut)
	assert.NotNil(t, gotCtx)
}

func TestGetAction(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		buf            bytes.Buffer
		isError        bool
		expectedResult Result
		actionFunction func(context.Context, io.Writer, *App, []string) error
		shor           shortener.Shortener
	}{
		{
			name:           "success",
			args:           []string{"Hpa3t2B"},
			actionFunction: getAction,
			buf:            bytes.Buffer{},
			expectedResult: Result{Code: "Hpa3t2B", Url: "https://example.com"},
			isError:        false,
			shor: &mockedShortener{
				getFunc: func(ctx context.Context, shortCode string) (string, error) {
					return "https://example.com", nil
				},
			},
		},
		{
			name:           "error produced",
			args:           []string{"Hpa3t2B"},
			actionFunction: getAction,
			buf:            bytes.Buffer{},
			isError:        true,
			shor: &mockedShortener{
				getFunc: func(ctx context.Context, url string) (string, error) {
					return "", errors.New("something went wrong")
				},
			},
		},
		{
			name:           "error empty args",
			args:           []string{},
			actionFunction: getAction,
			buf:            bytes.Buffer{},
			isError:        true,
			shor: &mockedShortener{
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

			app := &App{S: tc.shor}

			err := tc.actionFunction(ctx, &tc.buf, app, tc.args)

			if tc.isError {
				assert.Error(t, err)
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

func TestNewAGet(t *testing.T) {
	t.Cleanup(func() { getActionFunc = getAction })

	called := false
	var gotCtx context.Context
	var gotOut io.Writer
	var gotArgs []string

	getActionFunc = func(ctx context.Context, out io.Writer, a *App, args []string) error {
		called = true
		gotCtx = ctx
		gotOut = out
		gotArgs = append([]string(nil), args...)
		return nil
	}

	app := &App{S: &mockedShortener{}}
	cmd := NewGet(app)

	assert.Equal(t, "get <code>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"Hpa3t2B"})

	// Execute the command exactly like a user would
	require.NoError(t, cmd.ExecuteContext(context.Background()))

	// Assertions on wiring
	assert.True(t, called, "getActionFunc should be invoked")
	assert.Equal(t, []string{"Hpa3t2B"}, gotArgs)
	assert.Same(t, buf, gotOut)
	assert.NotNil(t, gotCtx)
}
