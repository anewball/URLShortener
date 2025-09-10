package cmd

import (
	"context"
	"log"
	"os"

	"github.com/anewball/urlshortener/internal/app"
	"github.com/spf13/cobra"
)

var (
	appInstance     *app.App
	actionsInstance Actions
)

type Result struct {
	ShortCode string `json:"shortCode"`
	RawURL    string `json:"rawUrl"`
}

type DeleteResponse struct {
	Deleted   bool   `json:"deleted"`
	ShortCode string `json:"shortCode"`
}

func NewRoot() *cobra.Command {
	var cfgFile string

	rootCmd := &cobra.Command{
		Use:           "urlshortener",
		Short:         "A simple URL shortener service",
		Long:          `A simple URL shortener service that allows you to shorten URLs and retrieve the original URLs using short codes.`,
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       "0.1.0",
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.urlshortener.yaml)")
	rootCmd.PersistentFlags().String("author", "Andy Newball", "author of the URL shortener")

	rootCmd.AddCommand(NewAdd(actionsInstance, appInstance.Shortener), NewDelete(actionsInstance, appInstance.Shortener), NewGet(actionsInstance, appInstance.Shortener), NewList(actionsInstance, appInstance.Shortener))

	return rootCmd
}

func Run(ctx context.Context, app *app.App, actions Actions, args ...string) error {
	log.SetOutput(os.Stderr)

	appInstance = app
	actionsInstance = actions

	defer func() {
		appInstance.Close()
		log.Println("Database connection pool closed")
	}()

	log.Println("Connected to database successfully")

	root := NewRoot()
	root.SetContext(ctx)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		return err
	}

	return nil
}
