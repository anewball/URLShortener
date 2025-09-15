package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/anewball/urlshortener/cmd"
	"github.com/anewball/urlshortener/config"
	"github.com/anewball/urlshortener/env"
	"github.com/anewball/urlshortener/internal/db"
	"github.com/anewball/urlshortener/internal/shortener"
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

	querier, err := db.NewQuerier(ctx, cfg)
	if err != nil {
		return err
	}
	log.Println("Connected to database successfully")
	defer func() {
		querier.Close()
		log.Println("Database connection pool closed")
	}()

	actions := cmd.NewActions()

	gen := shortener.NewNanoID(shortener.Alphabet)
	svc, err := shortener.New(querier, gen)
	if err != nil {
		return err
	}

	return cmd.Run(ctx, svc, actions)
}

func setupViper() map[string]string {
	_ = godotenv.Load()

	v := viper.New()
	v.AutomaticEnv()

	keys := []string{
		"POSTGRES_USER",
		"POSTGRES_PASSWORD",
		"POSTGRES_DB",
		"DB_MAX_CONNS",
		"DB_MIN_CONNS",
		"DB_MAX_CONN_LIFETIME",
		"DB_MAX_CONN_IDLE_TIME",
		"DB_URL",
	}

	for _, k := range keys {
		_ = v.BindEnv(k)
	}

	envMap := make(map[string]string, len(keys))
	for _, k := range keys {
		envMap[k] = strings.TrimSpace(v.GetString(k))
	}

	return envMap
}
