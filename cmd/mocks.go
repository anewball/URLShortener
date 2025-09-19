package cmd

import (
	"context"
	"io"

	"github.com/anewball/urlshortener/core"
)

var _ core.Actions = (*mockedActions)(nil)

type mockedActions struct {
	addActionFunc    func(ctx context.Context, out io.Writer, args []string) error
	getActionFunc    func(ctx context.Context, out io.Writer, args []string) error
	listActionFunc   func(ctx context.Context, limit int, offset int, out io.Writer) error
	deleteActionFunc func(ctx context.Context, out io.Writer, args []string) error
}

func (m *mockedActions) AddAction(ctx context.Context, out io.Writer, args []string) error {
	return m.addActionFunc(ctx, out, args)
}

func (m *mockedActions) GetAction(ctx context.Context, out io.Writer, args []string) error {
	return m.getActionFunc(ctx, out, args)
}

func (m *mockedActions) ListAction(ctx context.Context, limit int, offset int, out io.Writer) error {
	return m.listActionFunc(ctx, limit, offset, out)
}

func (m *mockedActions) DeleteAction(ctx context.Context, out io.Writer, args []string) error {
	return m.deleteActionFunc(ctx, out, args)
}
