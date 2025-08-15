package cmd

import (
	"context"

	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/jackc/pgx/v5/pgxpool"
)

type depsKey struct{}

type Deps struct {
	Pool      *pgxpool.Pool
	Shortener shortener.Shortener
}

func withDeps(ctx context.Context, d *Deps) context.Context {
	return context.WithValue(ctx, depsKey{}, d)
}

func getDeps(ctx context.Context) *Deps {
	v := ctx.Value(depsKey{})
	if v == nil {
		return nil
	}
	return v.(*Deps)
}
