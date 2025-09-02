package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/anewball/urlshortener/internal/app"
	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/spf13/cobra"
)

var addActionFunc = addAction // Indirection for testing

func NewAdd(app *app.App) *cobra.Command {
	return &cobra.Command{
		Use:   "add <url>",
		Short: "Save a URL to the shortener service",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return addActionFunc(cmd.Context(), cmd.OutOrStdout(), app.Shortener, args)
		},
	}
}

func addAction(ctx context.Context, out io.Writer, service shortener.Service, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(args) == 0 {
		return errors.New("requires at least 1 arg(s), only received 0")
	}

	arg := args[0]
	code, err := service.Add(ctx, arg)
	if err != nil {
		return fmt.Errorf("failed to add URL: %w", err)
	}

	result := Result{Code: code, Url: arg}

	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(result)
}
