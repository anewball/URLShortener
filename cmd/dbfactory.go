package cmd

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Factory defines behavior for building pools
type Factory interface {
	ParseConfig(dsn string) (*pgxpool.Config, error)
	NewWithConfig(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error)
}

// RealFactory is a factory that creates pools from a DSN
type RealFactory struct{}

func (f *RealFactory) ParseConfig(dsn string) (*pgxpool.Config, error) {
	return pgxpool.ParseConfig(dsn)
}

func (f *RealFactory) NewWithConfig(ctx context.Context, config *pgxpool.Config) (*pgxpool.Pool, error) {
	return pgxpool.NewWithConfig(ctx, config)
}
