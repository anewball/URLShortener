package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
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
	envMap := setupViper()

	en := env.New(envMap)
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

func setupViper() map[string]string {
	_ = godotenv.Load()
	v := viper.New()
	v.AutomaticEnv()

	_ = v.BindEnv("POSTGRES_USER")
	_ = v.BindEnv("POSTGRES_PASSWORD")
	_ = v.BindEnv("POSTGRES_DB")
	_ = v.BindEnv("DB_MAX_CONNS")
	_ = v.BindEnv("DB_MIN_CONNS")
	_ = v.BindEnv("DB_MAX_CONN_LIFETIME")
	_ = v.BindEnv("DB_MAX_CONN_IDLE_TIME")
	_ = v.BindEnv("DB_URL")

	envMap := map[string]string{}
	for _, key := range v.AllKeys() {
		k := strings.ToUpper(key)
		envMap[k] = strings.TrimSpace(v.GetString(key))
	}

	return envMap
}
