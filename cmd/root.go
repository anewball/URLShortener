package cmd

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/anewball/urlshortener/env"
	"github.com/anewball/urlshortener/internal/app"
	"github.com/anewball/urlshortener/internal/db"
	"github.com/spf13/cobra"
)

var (
	appInstance    *app.App
	newAppFunc     = app.New
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

	appInstance, err = newAppFunc(pool)
	if err != nil {
		return err
	}

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
	url, err := env.Get("DB_URL")
	if err != nil {
		return err
	}

	dbMaxConnsStr, err := env.Get("DB_MAX_CONNS")
	if err != nil {
		return err
	}
	dbMaxConns, err := strconv.Atoi(dbMaxConnsStr)
	if err != nil{
		return err
	}

	dbMinConnsStr, err := env.Get("DB_MIN_CONNS")
	if err != nil {
		return err
	}
	dbMinConns, err := strconv.Atoi(dbMinConnsStr)
	if err != nil {
		return err
	}

	dbMaxConnLifetimeStr, err := env.Get("DB_MAX_CONN_LIFETIME")
	if err != nil {
		return err
	}

	hour, err := time.ParseDuration(dbMaxConnLifetimeStr)
	if err != nil {
		return err
	}

	config := db.Config{
		URL:             url,
		MaxConns:        int32(dbMaxConns),
		MinConns:        int32(dbMinConns),
		MaxConnLifetime: hour,
	}

	return runWithFunc(ctx, config)
}
