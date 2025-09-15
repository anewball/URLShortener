package dbiface

import (
	"context"
)

type Querier interface {
	QueryRow(ctx context.Context, sql string, args ...any) Row
	Exec(ctx context.Context, sql string, arguments ...any) (CommandResult, error)
	Query(ctx context.Context, sql string, args ...any) (Rows, error)
	Close()
}

type Rows interface {
	Next() bool
	Scan(dest ...any) error
	Err() error
	Close()
}

type Row interface {
	Scan(dest ...any) error
}

type CommandResult interface {
	RowsAffected() int64
}
