package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/anewball/urlshortener/internal/app"
	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/spf13/cobra"
)

var listActionFunc = listAction

func NewList(app *app.App) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all URLs in the shortener service by offset and limit",
		RunE: func(cmd *cobra.Command, args []string) error {
			limit, _ := cmd.Flags().GetInt("limit")
			offset, _ := cmd.Flags().GetInt("offset")

			return listActionFunc(cmd.Context(), limit, offset, cmd.OutOrStdout(), app.Shortener)
		},
	}

	listCmd.Flags().IntP("limit", "n", 50, "max results to return")
	listCmd.Flags().IntP("offset", "o", 0, "results to skip")

	return listCmd
}

func listAction(ctx context.Context, limit int, offset int, out io.Writer, service shortener.Service) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if limit <= 0 {
		return fmt.Errorf("limit must be between 1 and 50")
	}
	if offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}

	urlItems, err := service.List(ctx, limit, offset)
	if err != nil {
		return fmt.Errorf("failed to list URLs: %w", err)
	}

	var results []Result = make([]Result, 0, len(urlItems))
	for _, u := range urlItems {
		results = append(results, Result{Code: u.ShortCode, Url: u.OriginalURL})
	}

	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(results)
}
