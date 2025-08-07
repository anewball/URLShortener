/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

// listCmd represents the list command
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all URLs in the shortener service by offset and limit",
	Aliases: []string{"l"},
	Args: cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error{
		offset, err := strconv.Atoi(args[0])
		if err != nil {
			return fmt.Errorf("invalid offset: %w", err)
		}

		limit, err := strconv.Atoi(args[1])
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		urls, err := cmdShortener.List(ctx, limit, offset)
		if err != nil {
			return fmt.Errorf("failed to list URLs: %w", err)
		}

		for _, url := range urls {
			fmt.Printf("http://localhost:8080/%s\n", url)
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
