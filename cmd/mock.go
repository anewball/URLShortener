package cmd

import (
	"context"

	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/jackc/pgx/v5/pgxpool"
)

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

type MockFactory struct {
	ParseConfigFunc   func(string) (*pgxpool.Config, error)
	NewWithConfigFunc func(context.Context, *pgxpool.Config) (*pgxpool.Pool, error)
}

func (m *MockFactory) ParseConfig(dns string) (*pgxpool.Config, error) {
	return m.ParseConfigFunc(dns)
}

func (m *MockFactory) NewWithConfig(ctx context.Context, cfg *pgxpool.Config) (*pgxpool.Pool, error) {
	return m.NewWithConfigFunc(ctx, cfg)
}
