package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/anewball/urlshortener/cmd"
	"github.com/anewball/urlshortener/config"
	"github.com/anewball/urlshortener/env"
	"github.com/anewball/urlshortener/internal/app"
	"github.com/anewball/urlshortener/internal/db"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	godotenv.Load()
	viper.AutomaticEnv()

	en := env.New()
	cfg, err := config.NewBuilder(en).FromEnv().Build()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	conn, err := db.NewPool(ctx, cfg)
	if err != nil {
		return err
	}

	app, err := app.New(conn)
	if err != nil {
		return err
	}

	actions := cmd.NewActions()

	return cmd.Run(ctx, app, actions)
}