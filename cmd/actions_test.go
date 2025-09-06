package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/anewball/urlshortener/env"
	"github.com/anewball/urlshortener/internal/app"
	"github.com/anewball/urlshortener/internal/db"
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
		actionFunction func(context.Context, io.Writer, shortener.URLShortener, []string) error
		shor           shortener.URLShortener
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

			err := tc.actionFunction(ctx, &tc.buf, tc.shor, tc.args)

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

	addActionFunc = func(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error {
		called = true
		gotCtx = ctx
		gotOut = out
		gotArgs = append([]string(nil), args...)
		return nil
	}

	cmd := NewAdd(&app.App{})

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
		actionFunction func(context.Context, io.Writer, shortener.URLShortener, []string) error
		shor           shortener.URLShortener
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

			err := tc.actionFunction(ctx, &tc.buf, tc.shor, tc.args)

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

func TestNewGet(t *testing.T) {
	t.Cleanup(func() { getActionFunc = getAction })

	called := false
	var gotCtx context.Context
	var gotOut io.Writer
	var gotArgs []string

	getActionFunc = func(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error {
		called = true
		gotCtx = ctx
		gotOut = out
		gotArgs = append([]string(nil), args...)
		return nil
	}

	cmd := NewGet(&app.App{})

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

func TestDeleteAction(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		buf            bytes.Buffer
		isError        bool
		expectedResult DeleteResponse
		actionFunction func(context.Context, io.Writer, shortener.URLShortener, []string) error
		shor           shortener.URLShortener
	}{
		{
			name:           "success",
			args:           []string{"Hpa3t2B"},
			actionFunction: deleteAction,
			buf:            bytes.Buffer{},
			expectedResult: DeleteResponse{Deleted: true, Code: "Hpa3t2B"},
			isError:        false,
			shor: &mockedShortener{
				deleteFunc: func(ctx context.Context, shortCode string) (bool, error) {
					return true, nil
				},
			},
		},
		{
			name:           "error produced",
			args:           []string{"Hpa3t2B"},
			actionFunction: deleteAction,
			buf:            bytes.Buffer{},
			isError:        true,
			shor: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, errors.New("something went wrong")
				},
			},
		},
		{
			name:           "error empty args",
			args:           []string{},
			actionFunction: deleteAction,
			buf:            bytes.Buffer{},
			isError:        true,
			shor: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, errors.New("requires at least 1 arg(s), only received 0")
				},
			},
		},
		{
			name:           "when row does not exists",
			args:           []string{"Hpa3t2B"},
			actionFunction: deleteAction,
			buf:            bytes.Buffer{},
			isError:        true,
			shor: &mockedShortener{
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

			err := tc.actionFunction(ctx, &tc.buf, tc.shor, tc.args)

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
	t.Cleanup(func() { deleteActionFunc = deleteAction })

	called := false
	var gotCtx context.Context
	var gotOut io.Writer
	var gotArgs []string

	deleteActionFunc = func(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error {
		called = true
		gotCtx = ctx
		gotOut = out
		gotArgs = append([]string(nil), args...)
		return nil
	}

	cmd := NewDelete(&app.App{})

	assert.Equal(t, "delete <code>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"Hpa3t2B"})

	// Execute the command exactly like a user would
	require.NoError(t, cmd.ExecuteContext(context.Background()))

	// Assertions on wiring
	assert.True(t, called, "deleteAction should be invoked")
	assert.Equal(t, []string{"Hpa3t2B"}, gotArgs)
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
		expectedResult []Result
		actionFunction func(context.Context, int, int, io.Writer, shortener.URLShortener) error
		shor           shortener.URLShortener
	}{
		{
			name:           "success",
			offset:         0,
			limit:          2,
			actionFunction: listAction,
			buf:            bytes.Buffer{},
			expectedResult: []Result{
				{Url: "https://anewball.com", Code: "nMHdgTh"},
				{Url: "https://jayden.newball.com", Code: "k5aBWD5"},
			},
			isError: false,
			shor: &mockedShortener{
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
			actionFunction: listAction,
			buf:            bytes.Buffer{},
			expectedResult: []Result{},
			isError:        true,
			shor:           &mockedShortener{},
		},
		{
			name:           "offset less than zero",
			offset:         -2,
			limit:          2,
			actionFunction: listAction,
			buf:            bytes.Buffer{},
			expectedResult: []Result{},
			isError:        true,
			shor:           &mockedShortener{},
		},
		{
			name:           "list returned error",
			offset:         0,
			limit:          2,
			actionFunction: listAction,
			buf:            bytes.Buffer{},
			expectedResult: []Result{},
			isError:        true,
			shor: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{}, errors.New("could not retrieve data")
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := tc.actionFunction(ctx, tc.limit, tc.offset, &tc.buf, tc.shor)

			if tc.isError {
				assert.Error(t, err)
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
	t.Cleanup(func() { listActionFunc = listAction })

	called := false
	var gotCtx context.Context
	var gotOut io.Writer

	listCmd := NewList(&app.App{})

	listActionFunc = func(ctx context.Context, limit int, offset int, out io.Writer, service shortener.URLShortener) error {
		called = true
		gotCtx = ctx
		gotOut = out
		return nil
	}

	assert.Equal(t, "list", listCmd.Use)
	assert.NotNil(t, listCmd.RunE)

	buf := &bytes.Buffer{}
	listCmd.SetOut(buf)
	listCmd.SetErr(io.Discard)
	listCmd.SetArgs([]string{"--offset", "0", "--limit", "2"})

	// Execute the command exactly like a user would
	require.NoError(t, listCmd.Execute())

	// Assertions on wiring
	assert.True(t, called, "deleteAction should be invoked")
	assert.Same(t, buf, gotOut)
	assert.NotNil(t, gotCtx)
}

func TestNewRoot(t *testing.T) {
	listCmd := NewRoot()

	assert.NotNil(t, listCmd)
}

func TestRunWith(t *testing.T) {
	// Restore after test
	t.Cleanup(func() { getNewPoolFunc = db.NewPool })
	t.Cleanup(func() { getActionFunc = getAction })

	testCases := []struct {
		name       string
		args       []string
		isError    bool
		newAppFunc func(pool db.Conn) (*app.App, error)
		dbFunc     func(ctx context.Context, cfg db.Config) (db.Conn, error)
		actionFunc func(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error
	}{
		{
			name:       "success",
			args:       []string{"get", "Hpa3t2B"},
			isError:    false,
			newAppFunc: newAppFunc,
			dbFunc: func(ctx context.Context, cfg db.Config) (db.Conn, error) {
				m := &mockPool{
					closeFunc: func() {
						log.Println("DB is closed")
					},
				}
				return m, nil
			},
			actionFunc: func(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error {
				return nil
			},
		},
		{
			name:       "failed when pool is nil",
			args:       []string{"get", "Hpa3t2B"},
			isError:    true,
			newAppFunc: newAppFunc,
			dbFunc: func(ctx context.Context, cfg db.Config) (db.Conn, error) {
				return nil, nil
			},
			actionFunc: func(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error {
				return nil
			},
		},
		{
			name:       "failed db",
			args:       []string{"get", "Hpa3t2B"},
			isError:    true,
			newAppFunc: newAppFunc,
			dbFunc: func(ctx context.Context, cfg db.Config) (db.Conn, error) {
				return nil, errors.New("error when opening db")
			},
			actionFunc: func(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error {
				return nil
			},
		},
		{
			name:       "failed with wrong command",
			args:       []string{"get1", "Hpa3t2B"},
			isError:    true,
			newAppFunc: newAppFunc,
			dbFunc: func(ctx context.Context, cfg db.Config) (db.Conn, error) {
				m := &mockPool{
					closeFunc: func() {
						log.Println("DB is closed")
					},
				}
				return m, nil
			},
			actionFunc: func(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error {
				return nil
			},
		},
		{
			name:       "defer db pool",
			args:       []string{"get", "Hpa3t2B"},
			isError:    false,
			newAppFunc: newAppFunc,
			dbFunc: func(ctx context.Context, cfg db.Config) (db.Conn, error) {
				m := &mockPool{
					closeFunc: func() {
						log.Println("DB is closed")
					},
				}
				return m, nil
			},
			actionFunc: func(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error {
				return nil
			},
		},
		{
			name:       "app error",
			args:       []string{"get", "Hpa3t2B"},
			isError:    true,
			newAppFunc: func(pool db.Conn) (*app.App, error) { return nil, errors.New("app error") },
			dbFunc: func(ctx context.Context, cfg db.Config) (db.Conn, error) {
				m := &mockPool{
					closeFunc: func() {
						log.Println("DB is closed")
					},
				}
				return m, nil
			},
			actionFunc: func(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error {
				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			getNewPoolFunc = tc.dbFunc
			getActionFunc = tc.actionFunc
			oldNewAppFunc := newAppFunc
			defer func() { newAppFunc = oldNewAppFunc }()

			newAppFunc = tc.newAppFunc

			err := runWithFunc(context.Background(), db.Config{}, tc.args...)

			if tc.isError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

func TestRun(t *testing.T) {
	oldNewEnv := newEnv
	t.Cleanup(func() {
		newEnv = oldNewEnv
		runWithFunc = runWith
	})

	testCases := []struct {
		name  string
		env   func() env.Env
		isErr bool
	}{
		{
			name:  "success",
			isErr: false,
			env: func() env.Env {
				return &mockEnv{
					getFunc: func(key string) (string, error) {
						data := map[string]string{
							"DB_URL":               "postgres://user:password@localhost:5432/your_db?sslmode=disable",
							"DB_MAX_CONNS":         "10",
							"DB_MIN_CONNS":         "1",
							"DB_MAX_CONN_LIFETIME": "1h",
						}
						return data[key], nil
					},
				}
			},
		},
		{
			name:  "DB_MAX_CONNS does not exist",
			isErr: true,
			env: func() env.Env {
				return &mockEnv{
					getFunc: func(key string) (string, error) {
						data := map[string]string{
							"DB_URL":               "postgres://user:password@localhost:5432/your_db?sslmode=disable",
							"DB_MAX_CONNS1":        "2",
							"DB_MIN_CONNS":         "1",
							"DB_MAX_CONN_LIFETIME": "1h",
						}
						if data[key] == "" {
							return "", errors.New("not found")
						}
						return data[key], nil
					},
				}
			},
		},
		{
			name:  "DB_MAX_CONNS not int",
			isErr: true,
			env: func() env.Env {
				return &mockEnv{
					getFunc: func(key string) (string, error) {
						data := map[string]string{
							"DB_URL":               "postgres://user:password@localhost:5432/your_db?sslmode=disable",
							"DB_MAX_CONNS":         "not-an-int",
							"DB_MIN_CONNS":         "1",
							"DB_MAX_CONN_LIFETIME": "1h",
						}
						return data[key], nil
					},
				}
			},
		},
		{
			name:  "DB_MIN_CONNS does not exist",
			isErr: true,
			env: func() env.Env {
				return &mockEnv{
					getFunc: func(key string) (string, error) {
						data := map[string]string{
							"DB_URL":               "postgres://user:password@localhost:5432/your_db?sslmode=disable",
							"DB_MAX_CONNS":         "2",
							"DB_MIN_CONNS1":        "1",
							"DB_MAX_CONN_LIFETIME": "1h",
						}
						if data[key] == "" {
							return "", errors.New("not found")
						}
						return data[key], nil
					},
				}
			},
		},
		{
			name:  "DB_MIN_CONNS not int",
			isErr: true,
			env: func() env.Env {
				return &mockEnv{
					getFunc: func(key string) (string, error) {
						data := map[string]string{
							"DB_URL":               "postgres://user:password@localhost:5432/your_db?sslmode=disable",
							"DB_MAX_CONNS":         "1",
							"DB_MIN_CONNS":         "not-an-int",
							"DB_MAX_CONN_LIFETIME": "1h",
						}
						return data[key], nil
					},
				}
			},
		},
		{
			name:  "DB_MAX_CONN_LIFETIME does not exist",
			isErr: true,
			env: func() env.Env {
				return &mockEnv{
					getFunc: func(key string) (string, error) {
						data := map[string]string{
							"DB_URL":                "postgres://user:password@localhost:5432/your_db?sslmode=disable",
							"DB_MAX_CONNS":          "2",
							"DB_MIN_CONNS":          "1",
							"DB_MAX_CONN_LIFETIME1": "1h",
						}
						if data[key] == "" {
							return "", errors.New("not found")
						}
						return data[key], nil
					},
				}
			},
		},
		{
			name:  "Error when Parsing DB_MAX_CONN_LIFETIME",
			isErr: true,
			env: func() env.Env {
				return &mockEnv{
					getFunc: func(key string) (string, error) {
						data := map[string]string{
							"DB_URL":               "postgres://user:password@localhost:5432/your_db?sslmode=disable",
							"DB_MAX_CONNS":         "2",
							"DB_MIN_CONNS":         "1",
							"DB_MAX_CONN_LIFETIME": "1p",
						}
						if data[key] == "" {
							return "", errors.New("not found")
						}
						return data[key], nil
					},
				}
			},
		},
		{
			name:  "failed",
			isErr: true,
			env: func() env.Env {
				return &mockEnv{
					getFunc: func(key string) (string, error) {
						return "", errors.New("not found")
					},
				}
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			runWithFunc = func(ctx context.Context, config db.Config, args ...string) error {
				return nil
			}

			newEnv = tc.env
			err := Run()
			if tc.isErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
