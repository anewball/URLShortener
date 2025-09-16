package cmd

import (
	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/spf13/cobra"
)

func NewDelete(acts Actions, svc shortener.URLShortener) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <code>",
		Short: "Delete a URL from the shortener service by short code",
		RunE: func(cmd *cobra.Command, args []string) error {
			return acts.DeleteAction(cmd.Context(), cmd.OutOrStdout(), svc, args)
		},
	}
}