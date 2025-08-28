package db

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	URL string
}

func NewPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, cfg.URL)
}
