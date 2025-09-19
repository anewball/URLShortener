package cmd

import (
	"github.com/anewball/urlshortener/core"
	"github.com/spf13/cobra"
)

func NewDelete(acts core.Actions) *cobra.Command {
	return &cobra.Command{
		Use:   "delete <code>",
		Short: "Delete a URL from the shortener service by short code",
		RunE: func(cmd *cobra.Command, args []string) error {
			return acts.DeleteAction(cmd.Context(), cmd.OutOrStdout(), args)
		},
	}
}
