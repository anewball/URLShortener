package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/spf13/cobra"
)

func NewList(a *App) *cobra.Command {
	c := &cobra.Command{
		Use:     "list",
		Short:   "List all URLs in the shortener service by offset and limit",
		Aliases: []string{"l"},
		RunE: func(cmd *cobra.Command, args []string) error {
			limit, _ := cmd.Flags().GetInt("limit")
			offset, _ := cmd.Flags().GetInt("offset")

			if limit <= 0 || offset < 0 {
				return fmt.Errorf("limit must be between 1 and 1000")
			}
			if offset < 0 {
				return fmt.Errorf("offset cannot be negative")
			}

			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			s := shortener.New(a.Pool)

			urlItems, err := s.List(ctx, limit, offset)
			if err != nil {
				return fmt.Errorf("failed to list URLs: %w", err)
			}

			for _, urlItem := range urlItems {
				fmt.Printf("http://localhost:8080/%s\n", urlItem.ShortCode)
			}

			return nil
		},
	}

	c.Flags().IntP("limit", "n", 50, "max results to return")
	c.Flags().IntP("offset", "o", 0, "results to skip")

	return c
}
