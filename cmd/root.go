package cmd

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:     "urlshortener",
	Short:   "A simple URL shortener service",
	Long:    `A simple URL shortener service that allows you to shorten URLs and retrieve the original URLs using short codes.`,
	Version: "0.1.0",
	PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
		initLoadEnv()
		dsn := os.Getenv("DATABASE_URL")
		if dsn == "" {
			return fmt.Errorf("DATABASE_URL environment variable is not set")
		}

		cfg, err := pgxpool.ParseConfig(dsn)
		if err != nil {
			log.Println("Unable to parse database connection string:", err)
			return err
		}

		cfg.MaxConns = 4                       // Set maximum number of connections to 4
		cfg.MinConns = 1                       // Set minimum number of connections to 1
		cfg.MaxConnLifetime = 30 * time.Minute // Set maximum connection lifetime to 30 minutes
		cfg.MaxConnIdleTime = 5 * time.Minute  // Set maximum idle time for connections to 5 minutes

		pool, err := pgxpool.NewWithConfig(cmd.Context(), cfg)
		if err != nil {
			log.Println("Unable to connect to database:", err)
			return err
		}

		deps := &Deps{Pool: pool, Shortener: shortener.NewShortener(pool)}
		cmd.SetContext(withDeps(cmd.Context(), deps))

		log.Println("Connected to database successfully")
		return nil
	},
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		if d := getDeps(cmd.Context()); d != nil && d.Pool != nil {
			d.Pool.Close()
			log.Println("Database connection pool closed")
		}
	},
}

func Execute() error {
	if err := rootCmd.Execute(); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.urlshortener.yaml)")
	rootCmd.PersistentFlags().StringP("author", "", "Andy Newball", "author of the URL shortener")

	listCmd.Flags().IntP("limit", "n", 50, "max results to return")
	listCmd.Flags().IntP("offset", "o", 0, "results to skip")
}

func initLoadEnv() {
	// Load environment variables from .env file if needed
	// This can be done using a package like godotenv
	_ = godotenv.Load(".env")
}
