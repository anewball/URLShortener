package cmd

import (
	"bytes"
	"context"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAdd(t *testing.T) {
	called := false
	var gotCtx context.Context
	var gotOut io.Writer
	var gotArgs []string

	mActions := &mockedActions{
		addActionFunc: func(ctx context.Context, out io.Writer, args []string) error {
			called = true
			gotCtx = ctx
			gotOut = out
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	cmd := NewAdd(mActions)

	assert.Equal(t, "add <url>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	args := []string{"https://example.com"}

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)

	// Execute the command exactly like a user would
	require.NoError(t, cmd.ExecuteContext(context.Background()))

	// Assertions on wiring
	assert.True(t, called, "AddAction should be invoked")
	assert.Equal(t, args, gotArgs)
	assert.Same(t, buf, gotOut)
	assert.NotNil(t, gotCtx)
}

func TestNewGet(t *testing.T) {
	called := false
	var gotCtx context.Context
	var gotOut io.Writer
	var gotArgs []string

	mActions := &mockedActions{
		getActionFunc: func(ctx context.Context, out io.Writer, args []string) error {
			called = true
			gotCtx = ctx
			gotOut = out
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	cmd := NewGet(mActions)

	assert.Equal(t, "get <code>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	args := []string{"Hpa3t2B"}

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)

	// Execute the command exactly like a user would
	require.NoError(t, cmd.ExecuteContext(context.Background()))

	// Assertions on wiring
	assert.True(t, called, "GetAction should be invoked")
	assert.Equal(t, args, gotArgs)
	assert.Same(t, buf, gotOut)
	assert.NotNil(t, gotCtx)
}

func TestNewDelete(t *testing.T) {
	called := false
	var gotCtx context.Context
	var gotOut io.Writer
	var gotArgs []string

	mActions := &mockedActions{
		deleteActionFunc: func(ctx context.Context, out io.Writer, args []string) error {
			called = true
			gotCtx = ctx
			gotOut = out
			gotArgs = append([]string(nil), args...)
			return nil
		},
	}

	cmd := NewDelete(mActions)

	assert.Equal(t, "delete <code>", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	args := []string{"Hpa3t2B"}

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs(args)

	// Execute the command exactly like a user would
	require.NoError(t, cmd.ExecuteContext(context.Background()))

	// Assertions on wiring
	assert.True(t, called, "DeleteAction should be invoked")
	assert.Equal(t, args, gotArgs)
	assert.Same(t, buf, gotOut)
	assert.NotNil(t, gotCtx)
}

func TestNewList(t *testing.T) {
	called := false
	var gotCtx context.Context
	var gotOut io.Writer

	mActions := &mockedActions{
		listActionFunc: func(ctx context.Context, limit int, offset int, out io.Writer) error {
			called = true
			gotCtx = ctx
			gotOut = out
			return nil
		},
	}

	cmd := NewList(mActions)

	assert.Equal(t, "list", cmd.Use)
	assert.NotNil(t, cmd.RunE)

	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(io.Discard)
	cmd.SetArgs([]string{"--offset", "0", "--limit", "2"})

	// Execute the command exactly like a user would
	require.NoError(t, cmd.Execute())

	// Assertions on wiring
	assert.True(t, called, "ListAction should be invoked")
	assert.Same(t, buf, gotOut)
	assert.NotNil(t, gotCtx)
}

func TestNewRoot(t *testing.T) {
	cmd := NewRoot(&mockedActions{})

	assert.Equal(t, "urlshortener", cmd.Use)
}
