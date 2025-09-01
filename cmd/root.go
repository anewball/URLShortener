package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anewball/urlshortener/env"
	"github.com/anewball/urlshortener/internal/app"
	"github.com/anewball/urlshortener/internal/db"
	"github.com/spf13/cobra"
)

var (
	appInstance    *app.App
	getNewPoolFunc = db.NewPool
	newEnv         = env.New
	runWithFunc    = runWith
)

type Result struct {
	Code string `json:"code"`
	Url  string `json:"url"`
}

type DeleteResponse struct {
	Deleted bool   `json:"deleted"`
	Code    string `json:"code"`
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

	rootCmd.AddCommand(NewAdd(appInstance), NewDelete(appInstance), NewGet(appInstance), NewList(appInstance))

	return rootCmd
}

func runWith(ctx context.Context, config db.Config, args ...string) error {
	log.SetOutput(os.Stderr)

	pool, err := getNewPoolFunc(ctx, config)
	if err != nil {
		return err
	}

	if pool == nil {
		return fmt.Errorf("db: nil pool")
	}

	appInstance = app.New(pool)

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

func Run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	env := newEnv()
	URL, err := env.Get("db.url")
	if err != nil {
		return err
	}

	config := db.Config{
		URL:             URL,
		MaxConns:        5,
		MinConns:        1,
		MaxConnLifetime: time.Hour,
	}

	return runWithFunc(ctx, config)
}
