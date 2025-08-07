package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
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
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		panic(err)
	}
}

func init() {
	cobra.OnInitialize(
		initLoadEnv,
		initConfig,
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

	dbConnectionString = os.Getenv("DB_CONNECTION_STRING")
}

func initConfig() {
	// Configuration initialization logic can go here if needed
	// For example, reading from a config file or environment variables
	ctx, stop := context.WithCancel(context.Background())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	var err error
	pool, err = pgxpool.New(ctx, dbConnectionString)
	if err != nil {
		log.Fatal("Unable to connect to database:", err)
	}

	cmdShortener = initializeShortener(pool)

	// Goroutine to keep the pool alive
	go func() {
		<-c
		log.Println("Shutting down gracefully...")

		stop()

		time.Sleep(2 * time.Second)
		log.Println("Exiting...")
		os.Exit(0)
	}()
}

func initializeShortener(pool *pgxpool.Pool) shortener.Shortener{
	if pool == nil {
		log.Fatal("Database connection pool is not initialized")
	}
	return shortener.NewShortener(pool)
}