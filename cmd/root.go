package cmd

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/anewball/urlshortener/internal/db"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var pool *pgxpool.Pool

func getDBPool(ctx context.Context) (*pgxpool.Pool, error) {
	dbCfg := db.Config{
		URL: viper.GetString("db.url"),
	}

	return db.NewPool(ctx, dbCfg)
}

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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			p, err := getDBPool(cmd.Context())
			if err != nil {
				return err
			}

			pool = p

			return nil
		},
	}

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.urlshortener.yaml)")
	rootCmd.PersistentFlags().String("author", "Andy Newball", "author of the URL shortener")

	rootCmd.AddCommand(NewAdd(), NewDelete(), NewGet(), NewList())

	return rootCmd
}

func runWith(ctx context.Context, args ...string) error {
	log.SetOutput(os.Stderr)

	defer func() {
		pool.Close()
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

	return runWith(ctx)
}
