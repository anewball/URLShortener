package db

import (
	"context"
	"fmt"

	"github.com/anewball/urlshortener/config"
	"github.com/anewball/urlshortener/internal/dbiface"
	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewQuerier(ctx context.Context, cfg config.Config) (dbiface.Querier, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("db: empty connection URL")
	}

	// Parse configuration from URL
	poolConfig, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("db: invalid connection URL: %w", err)
	}

	// Optional tuning
	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		poolConfig.MinConns = cfg.MinConns
	}
	if cfg.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	}

	// Create pool
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("db: failed to create pool: %w", err)
	}

	// Verify connection with a ping
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: ping failed: %w", err)
	}
	return &poolAdapter{pool}, nil
}

type rowsAdapter struct{ pgx.Rows }

func (r *rowsAdapter) Close() { r.Rows.Close() }

type rowAdapter struct{ pgx.Row }

func (r rowAdapter) Scan(dest ...any) error {
	if err := r.Row.Scan(dest...); err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("%w: %w", shortener.ErrNotFound, err)
		}
		return err
	}
	return nil
}

type poolAdapter struct{ *pgxpool.Pool }

type commandTagAdapter struct{ tag pgconn.CommandTag }

func (c commandTagAdapter) RowsAffected() int64 {
	return c.tag.RowsAffected()
}

func (p *poolAdapter) QueryRow(ctx context.Context, sql string, args ...any) dbiface.Row {
	return &rowAdapter{p.Pool.QueryRow(ctx, sql, args...)}
}

func (p *poolAdapter) Exec(ctx context.Context, sql string, args ...any) (dbiface.CommandResult, error) {
	tag, err := p.Pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return commandTagAdapter{tag: tag}, nil
}

func (p *poolAdapter) Query(ctx context.Context, sql string, args ...any) (dbiface.Rows, error) {
	rows, err := p.Pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &rowsAdapter{rows}, nil
}

func (p *poolAdapter) Close() {
	p.Pool.Close()
}
