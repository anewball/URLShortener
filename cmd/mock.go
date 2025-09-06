package cmd

import (
	"context"

	"github.com/anewball/urlshortener/env"
	"github.com/anewball/urlshortener/internal/db"
	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

var _ shortener.URLShortener = (*mockedShortener)(nil)

type mockedShortener struct {
	addFunc    func(ctx context.Context, url string) (string, error)
	getFunc    func(ctx context.Context, shortCode string) (string, error)
	listFunc   func(ctx context.Context, limit, offset int) ([]shortener.URLItem, error)
	deleteFunc func(ctx context.Context, shortCode string) (bool, error)
}

func (m *mockedShortener) Add(ctx context.Context, url string) (string, error) {
	return m.addFunc(ctx, url)
}

func (m *mockedShortener) Get(ctx context.Context, code string) (string, error) {
	return m.getFunc(ctx, code)
}

func (m *mockedShortener) List(ctx context.Context, limit, offset int) ([]shortener.URLItem, error) {
	return m.listFunc(ctx, limit, offset)
}

func (m *mockedShortener) Delete(ctx context.Context, code string) (bool, error) {
	return m.deleteFunc(ctx, code)
}

var _ db.Conn = (*mockPool)(nil)

type mockPool struct {
	queryRowFunc func(ctx context.Context, sql string, args ...any) pgx.Row
	execFunc     func(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	queryFunc    func(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	closeFunc    func()
}

func (m *mockPool) QueryRow(ctx context.Context, sql string, args ...any) pgx.Row {
	if m.queryRowFunc != nil {
		return m.queryRowFunc(ctx, sql, args...)
	}
	return nil
}

func (m *mockPool) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if m.execFunc != nil {
		return m.execFunc(ctx, sql, args...)
	}
	return pgconn.CommandTag{}, nil
}

func (m *mockPool) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if m.queryFunc != nil {
		return m.queryFunc(ctx, sql, args...)
	}
	return nil, nil
}

func (m *mockPool) Close() {
	if m.closeFunc != nil {
		m.closeFunc()
	}
}

var _ env.Env = (*mockEnv)(nil)

type mockEnv struct {
	getFunc func(key string) (string, error)
}

func (m *mockEnv) Get(key string) (string, error) {
	if m.getFunc != nil {
		return m.getFunc(key)
	}
	return "", nil
}
