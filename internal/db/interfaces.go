package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	Close()
}

var _ Querier = (*pgxpool.Pool)(nil)

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}

type Row interface {
	Scan(dest ...any) error
}
