package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"anewball.com/internal/shortener"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/spf13/cobra"
)

var cfgFile string
var dbConnectionString string
var cmdShortener shortener.Shortener
var pool *pgxpool.Pool

var rootCmd = &cobra.Command{
	Use:     "urlshortener",
	Short:   "A simple URL shortener service",
	Long:    `A simple URL shortener service that allows you to shorten URLs and retrieve the original URLs using short codes.`,
	Version: "0.1.0",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if dbConnectionString == "" {
			return cmd.Help()
		}

		ctx, cancel := context.WithTimeout(cmd.Context(), 20*time.Second)
		defer cancel()

		var err error
		pool, err = pgxpool.New(ctx, dbConnectionString)
		if err != nil {
			log.Fatal("Unable to connect to database:", err)
			return err
		}

		cmdShortener = initializeShortener(pool)
		if cmdShortener == nil {
			return fmt.Errorf("failed to initialize shortener")
		}

		log.Println("Connected to database successfully")
		return nil
	},
	PersistentPostRunE: func(cmd *cobra.Command, args []string) error {
		if pool != nil {
			pool.Close()
			log.Println("Database connection pool closed")
		}
		return nil
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	cobra.OnInitialize(
		initLoadEnv,
	)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.urlshortener.yaml)")
	rootCmd.PersistentFlags().StringP("author", "", "Andy Newball", "author of the URL shortener")
}

func initLoadEnv() {
	// Load environment variables from .env file if needed
	// This can be done using a package like godotenv
	if err := godotenv.Load("cmd/config/.env"); err != nil {
		log.Print("No .env file found")
	}

	dbConnectionString = os.Getenv("DATABASE_URL")
}

func initializeShortener(pool *pgxpool.Pool) shortener.Shortener {
	if pool == nil {
		log.Fatal("Database connection pool is not initialized")
	}
	return shortener.NewShortener(pool)
}
