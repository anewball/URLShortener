package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/spf13/cobra"
)

func NewAdd(a *App) *cobra.Command {
	return &cobra.Command{
		Use:   "add <url>",
		Short: "Save a URL to the shortener service",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			return addAction(ctx, cmd.OutOrStdout(), a, args)
		},
	}
}

func addAction(ctx context.Context, out io.Writer, a *App, args []string) error {
	s := shortener.New(a.Pool)

	arg := args[0]
	code, err := s.Add(ctx, arg)
	if err != nil {
		return fmt.Errorf("failed to add URL: %w", err)
	}

	result := Result{Code: code, Url: arg}

	encoder := json.NewEncoder(out)

	return encoder.Encode(result)
}
