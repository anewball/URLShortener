package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

type App struct {
	Pool *pgxpool.Pool
}

func NewRoot(a *App) *cobra.Command {
	var cfgFile string

	rootCmd := &cobra.Command{
		Use:     "urlshortener",
		Short:   "A simple URL shortener service",
		Long:    `A simple URL shortener service that allows you to shorten URLs and retrieve the original URLs using short codes.`,
		Version: "0.1.0",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// Load environment variables from .env file if needed
			// This can be done using a package like godotenv
			_ = godotenv.Load()

			dsn := os.Getenv("DATABASE_URL")
			if dsn == "" {
				return fmt.Errorf("DATABASE_URL environment variable is not set")
			}

			p, err := newPool(cmd.Context(), dsn)
			if err != nil {
				return err
			}

			a.Pool = p

			log.Println("Connected to database successfully")
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if a.Pool != nil {
				a.Pool.Close()
				a.Pool = nil
				log.Println("Database connection pool closed")
			}
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.urlshortener.yaml)")
	rootCmd.PersistentFlags().String("author", "Andy Newball", "author of the URL shortener")

	rootCmd.AddCommand(NewAdd(a), NewDelete(a), NewGet(a), NewList(a))

	return rootCmd
}

func newPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = 4                       // Set maximum number of connections to 4
	cfg.MinConns = 1                       // Set minimum number of connections to 1
	cfg.MaxConnLifetime = 30 * time.Minute // Set maximum connection lifetime to 30 minutes
	cfg.MaxConnIdleTime = 5 * time.Minute  // Set maximum idle time for connections to 5 minutes
	cfg.HealthCheckPeriod = 30 * time.Second

	return pgxpool.NewWithConfig(ctx, cfg)
}
