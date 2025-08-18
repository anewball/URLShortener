package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/spf13/cobra"
)

func NewDelete(a *App) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <code>",
		Short: "Delete a URL from the shortener service by short code",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("url code is required")
			}
			code := args[0]

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			s := shortener.New(a.Pool)

			deleted, err := s.Delete(ctx, code)
			if err != nil {
				return fmt.Errorf("failed to delete URL: %w", err)
			}
			if !deleted {
				return fmt.Errorf("no URL found for code %q", code)
			}

			fmt.Println("URL deleted successfully")
			return nil
		},
	}
}
