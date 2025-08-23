package cmd

import (
	"bytes"
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/stretchr/testify/assert"
)

func TestActions(t *testing.T) {
	testCases := []struct {
		name           string
		args           []string
		out            io.Writer
		isError        bool
		actionFunction func(context.Context, io.Writer, *App, []string) error
		shor           shortener.Shortener
	}{
		{
			name:           "success",
			args:           []string{"https://example.com"},
			actionFunction: addAction,
			out:            &bytes.Buffer{},
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
			out:            &bytes.Buffer{},
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

			err := tc.actionFunction(ctx, tc.out, app, tc.args)

			if !tc.isError {
				assert.Nil(t, err)
			} else {
				assert.NotNil(t, err)
			}
		})
	}
}
