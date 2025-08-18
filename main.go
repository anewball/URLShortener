package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/anewball/urlshortener/cmd"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	a := &cmd.App{}
	root := cmd.NewRoot(a)
	root.SetContext(ctx)

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
