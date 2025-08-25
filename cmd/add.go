package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/spf13/cobra"
)

func NewAdd(a *App) *cobra.Command {
	return &cobra.Command{
		Use:   "add <url>",
		Short: "Save a URL to the shortener service",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return addAction(cmd.Context(), cmd.OutOrStdout(), a, args)
		},
	}
}

func addAction(ctx context.Context, out io.Writer, a *App, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(args) == 0 {
		return errors.New("requires at least 1 arg(s), only received 0")
	}

	arg := args[0]
	code, err := a.S.Add(ctx, arg)
	if err != nil {
		return fmt.Errorf("failed to add URL: %w", err)
	}

	result := Result{Code: code, Url: arg}

	encoder := json.NewEncoder(out)

	return encoder.Encode(result)
}
