package cmd

import (
	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/spf13/cobra"
)

func NewGet(acts Actions, svc shortener.URLShortener) *cobra.Command {
	return &cobra.Command{
		Use:   "get <code>",
		Short: "Retrieve a URL from the shortener service",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return acts.GetAction(cmd.Context(), cmd.OutOrStdout(), svc, args)
		},
	}
}