package cmd

import (
	"github.com/anewball/urlshortener/core"
	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/spf13/cobra"
)

func NewList(acts core.Actions, svc shortener.URLShortener) *cobra.Command {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List all URLs in the shortener service by offset and limit",
		Example: `
		  	urlshortener list --offset 0 --limit 10
  			urlshortener list -o 0 -n 10`,
		RunE: func(cmd *cobra.Command, args []string) error {
			limit, _ := cmd.Flags().GetInt("limit")
			offset, _ := cmd.Flags().GetInt("offset")

			return acts.ListAction(cmd.Context(), limit, offset, cmd.OutOrStdout(), svc)
		},
	}

	listCmd.Flags().IntP("limit", "n", 50, "max results to return")
	listCmd.Flags().IntP("offset", "o", 0, "results to skip")

	return listCmd
}
