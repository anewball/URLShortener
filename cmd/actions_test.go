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
)

func TestActions(t *testing.T) {
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
