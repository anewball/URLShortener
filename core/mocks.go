package core

import (
	"context"

	"github.com/anewball/urlshortener/internal/shortener"
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
