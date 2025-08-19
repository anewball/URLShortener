package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/spf13/cobra"
)

func NewGet(a *App) *cobra.Command {
	c := &cobra.Command{
		Use:   "get <code>",
		Short: "Retrieve a URL from the shortener service",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
			defer cancel()

			s := shortener.New(a.Pool)

			url, err := s.Get(ctx, args[0])
			if err != nil {
				return fmt.Errorf("failed to retrieve original URL: %w", err)
			}

			result := Result{Code: args[0], Url: url}
			encoder := json.NewEncoder(cmd.OutOrStdout())
			
			return encoder.Encode(result)
		},
	}
	return c
}
