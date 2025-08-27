package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
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

func newPool(ctx context.Context, dns string, factory Factory) (*pgxpool.Pool, error) {
	cfg, err := factory.ParseConfig(dns)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = 4                       // Set maximum number of connections to 4
	cfg.MinConns = 1                       // Set minimum number of connections to 1
	cfg.MaxConnLifetime = 30 * time.Minute // Set maximum connection lifetime to 30 minutes
	cfg.MaxConnIdleTime = 5 * time.Minute  // Set maximum idle time for connections to 5 minutes
	cfg.HealthCheckPeriod = 30 * time.Second

	return factory.NewWithConfig(ctx, cfg)
}

func Run() error {
	log.SetOutput(os.Stderr)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Load environment variables from .env file if needed
	// This can be done using a package like godotenv
	_ = godotenv.Load()

	var env Env = &realEnv{}

	dsn := env.Get("DATABASE_URL")
	if dsn == "" {
		return fmt.Errorf("DATABASE_URL environment variable is not set")
	}

	var factory Factory = &RealFactory{}

	p, err := newPool(ctx, dsn, factory)
	if err != nil {
		return err
	}
	log.Println("Connected to database successfully")

	app := NewApp(p)

	root := NewRoot(app)
	root.SetContext(ctx)

	if err := root.Execute(); err != nil {
		return err
	}

	defer p.Close()
	log.Println("Database connection pool closed")

	return nil
}
