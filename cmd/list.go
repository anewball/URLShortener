package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List all URLs in the shortener service by offset and limit",
	Aliases: []string{"l"},
	RunE: func(cmd *cobra.Command, args []string) error {
		deps := getDeps(cmd.Context())
		if deps == nil {
			return fmt.Errorf("internal: deps not set")
		}

		limit, _ := cmd.Flags().GetInt("limit")
		offset, _ := cmd.Flags().GetInt("offset")

		if limit <= 0 || offset < 0 {
			return fmt.Errorf("limit must be between 1 and 1000")
		}
		if offset < 0 {
			return fmt.Errorf("offset cannot be negative")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		urlItems, err := deps.Shortener.List(ctx, limit, offset)
		if err != nil {
			return fmt.Errorf("failed to list URLs: %w", err)
		}

		for _, urlItem := range urlItems {
			fmt.Printf("http://localhost:8080/%s\n", urlItem.ShortCode)
		}

		return nil
	},
}

func init() {
	rootCmd.AddCommand(listCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// listCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// listCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
