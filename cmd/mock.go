package cmd

import (
	"context"
	"io"

	"github.com/anewball/urlshortener/core"
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

var _ core.Actions = (*mockedActions)(nil)

type mockedActions struct {
	AddActionFunc    func(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error
	GetActionFunc    func(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error
	ListActionFunc   func(ctx context.Context, limit int, offset int, out io.Writer, svc shortener.URLShortener) error
	DeleteActionFunc func(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error
}

func (m *mockedActions) AddAction(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error {
	return m.AddActionFunc(ctx, out, svc, args)
}

func (m *mockedActions) GetAction(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error {
	return m.GetActionFunc(ctx, out, svc, args)
}

func (m *mockedActions) ListAction(ctx context.Context, limit int, offset int, out io.Writer, svc shortener.URLShortener) error {
	return m.ListActionFunc(ctx, limit, offset, out, svc)
}

func (m *mockedActions) DeleteAction(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error {
	return m.DeleteActionFunc(ctx, out, svc, args)
}
