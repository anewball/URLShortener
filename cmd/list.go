package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

func NewList(a *App) *cobra.Command {
	c := &cobra.Command{
		Use:   "list",
		Short: "List all URLs in the shortener service by offset and limit",
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

			urlItems, err := a.S.List(ctx, limit, offset)
			if err != nil {
				return fmt.Errorf("failed to list URLs: %w", err)
			}

			var results []Result
			for _, u := range urlItems {
				results = append(results, Result{Code: u.ShortCode, Url: u.OriginalURL})
			}

			encoder := json.NewEncoder(cmd.OutOrStdout())

			return encoder.Encode(results)
		},
	}

	c.Flags().IntP("limit", "n", 50, "max results to return")
	c.Flags().IntP("offset", "o", 0, "results to skip")

	return c
}
