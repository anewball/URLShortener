package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/spf13/cobra"
)

var getActionFunc = getAction

func NewGet() *cobra.Command {
	return &cobra.Command{
		Use:   "get <code>",
		Short: "Retrieve a URL from the shortener service",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			service := shortener.New(pool)
			return getActionFunc(cmd.Context(), cmd.OutOrStdout(), service, args)
		},
	}
}

func getAction(ctx context.Context, out io.Writer, service shortener.Shortener, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(args) == 0 {
		return errors.New("requires at least 1 arg(s), only received 0")
	}

	arg := args[0]
	url, err := service.Get(ctx, arg)
	if err != nil {
		return fmt.Errorf("failed to retrieve original URL: %w", err)
	}

	result := Result{Code: arg, Url: url}
	encoder := json.NewEncoder(out)

	return encoder.Encode(result)
}
