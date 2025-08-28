package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PoolFactory interface {
	NewPool(context.Context, string) (*pgxpool.Pool, error)
}

type PostgresPoolFactory struct{}

func (p *PostgresPoolFactory) NewPool(ctxt context.Context, dsn string) (*pgxpool.Pool, error) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}

	cfg.MaxConns = 4                       // Set maximum number of connections to 4
	cfg.MinConns = 1                       // Set minimum number of connections to 1
	cfg.MaxConnLifetime = 30 * time.Minute // Set maximum connection lifetime to 30 minutes
	cfg.MaxConnIdleTime = 5 * time.Minute  // Set maximum idle time for connections to 5 minutes
	cfg.HealthCheckPeriod = 30 * time.Second

	return pgxpool.NewWithConfig(ctx, cfg)
}
