package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
)

var listActionFunc = listAction

func NewList(app *App) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List all URLs in the shortener service by offset and limit",
		RunE: func(cmd *cobra.Command, args []string) error {
			limit, _ := cmd.Flags().GetInt("limit")
			offset, _ := cmd.Flags().GetInt("offset")

			return listActionFunc(cmd.Context(), limit, offset, cmd.OutOrStdout(), app)
		},
	}
}

func listAction(ctx context.Context, limit int, offset int, out io.Writer, app *App) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if limit <= 0 {
		return fmt.Errorf("limit must be between 1 and 1000")
	}
	if offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}

	urlItems, err := app.S.List(ctx, limit, offset)
	if err != nil {
		return fmt.Errorf("failed to list URLs: %w", err)
	}

	var results []Result
	for _, u := range urlItems {
		results = append(results, Result{Code: u.ShortCode, Url: u.OriginalURL})
	}

	encoder := json.NewEncoder(out)

	return encoder.Encode(results)
}
