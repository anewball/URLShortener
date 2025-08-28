package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/spf13/cobra"
)

type App struct {
	S shortener.Shortener
}

func NewApp(dbConn shortener.DatabaseConn) *App {
	return &App{S: shortener.New(dbConn)}
}

type Result struct {
	Code string `json:"code"`
	Url  string `json:"url"`
}

type DeleteResponse struct {
	Deleted bool   `json:"deleted"`
	Code    string `json:"code"`
}

func NewRoot(app *App) *cobra.Command {
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

	rootCmd.AddCommand(NewAdd(app), NewDelete(app), NewGet(app), NewList(app))

	return rootCmd
}

func runWith(ctx context.Context, env Env, poolFactory PoolFactory, app *App, args ...string) error {
	log.SetOutput(os.Stderr)

	if app != nil && app.S != nil {
		root := NewRoot(app)
		root.SetContext(ctx)
		root.SetArgs(args)
		return root.Execute()
	}

	dsn := env.Get("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	pool, err := poolFactory.NewPool(ctx, dsn)
	if err != nil {
		return err
	}
	defer func() {
		pool.Close()
		log.Println("Database connection pool closed")
	}()

	log.Println("Connected to database successfully")

	app.S = shortener.New(pool)

	root := NewRoot(app)
	root.SetContext(ctx)
	root.SetArgs(args)

	if err := root.Execute(); err != nil {
		return err
	}

	return nil
}

func Run() error {
	app := &App{}
	var env Env = &realEnv{}
	var poolFactory PoolFactory = &PostgresPoolFactory{}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	return runWith(ctx, env, poolFactory, app)
}
