package cmd

import (
	"github.com/anewball/urlshortener/core"
	"github.com/spf13/cobra"
)

func NewGet(acts core.Actions) *cobra.Command {
	return &cobra.Command{
		Use:   "get <code>",
		Short: "Retrieve a URL from the shortener service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return acts.GetAction(cmd.Context(), cmd.OutOrStdout(), args)
		},
	}
}
