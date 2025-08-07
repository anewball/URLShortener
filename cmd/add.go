/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// addCmd represents the add command
var addCmd = &cobra.Command{
	Use:          "add [url]",
	Short:        "Save a URL to the shortener service",
	Aliases:      []string{"a"},
	Args:         cobra.MinimumNArgs(1),
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		url := args[0]

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		urlCode, err := cmdShortener.Add(ctx, url)
		if err != nil {
			return fmt.Errorf("failed to add URL: %w", err)
		}
		fmt.Printf("Shortened URL: %s/%s\n", "http://localhost:8080", urlCode)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(addCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// addCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// addCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}