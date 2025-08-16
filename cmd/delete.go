package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete <code>",
	Short: "Delete a URL from the shortener service by short code",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return fmt.Errorf("url code is required")
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 5*time.Second)
		defer cancel()

		code := args[0]
		deleted, err := app.Short.Delete(ctx, code)
		if err != nil {
			return fmt.Errorf("failed to delete URL: %w", err)
		}
		if !deleted{
			return fmt.Errorf("no URL found for code %q", code)
		}

		fmt.Println("URL deleted successfully")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// deleteCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// deleteCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
