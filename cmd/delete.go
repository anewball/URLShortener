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

var deleteActionFunc = deleteAction

func NewDelete() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <code>",
		Short: "Delete a URL from the shortener service by short code",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := shortener.New(pool)
			return deleteActionFunc(cmd.Context(), cmd.OutOrStdout(), service, args)
		},
	}
}

func deleteAction(ctx context.Context, out io.Writer, service shortener.Shortener, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(args) == 0 {
		return fmt.Errorf("requires at least 1 arg(s), only received 0")
	}
	code := args[0]

	deleted, err := service.Delete(ctx, code)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}
	if !deleted {
		return fmt.Errorf("no URL found for code %q", code)
	}

	var response DeleteResponse

	response.Deleted = deleted
	response.Code = code

	encoder := json.NewEncoder(out)

	return encoder.Encode(response)
}
