package cmd

import (
	"github.com/anewball/urlshortener/core"
	"github.com/spf13/cobra"
)

func NewAdd(acts core.Actions) *cobra.Command {
	return &cobra.Command{
		Use:   "add <url>",
		Short: "Save a URL to the shortener service",
		RunE: func(cmd *cobra.Command, args []string) error {
			return acts.AddAction(cmd.Context(), cmd.OutOrStdout(), args)
		},
	}
}
