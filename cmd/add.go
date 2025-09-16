package cmd

import (
	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/spf13/cobra"
)

func NewAdd(acts Actions, svc shortener.URLShortener) *cobra.Command {
	return &cobra.Command{
		Use:   "add <url>",
		Short: "Save a URL to the shortener service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return acts.AddAction(cmd.Context(), cmd.OutOrStdout(), svc, args)
		},
	}
}