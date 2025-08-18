package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/spf13/cobra"
)

func NewAdd(a *App) *cobra.Command {
	return &cobra.Command{
		Use:          "add <url>",
		Short:        "Save a URL to the shortener service",
		Args:         cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			s := shortener.New(a.Pool)

			code, err := s.Add(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to add URL: %w", err)
			}
			fmt.Printf("Shortened URL: %s/%s\n", "http://localhost:8080", code)

			return nil
		},
	}
}
