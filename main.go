package main

import (
	"os"

	"github.com/anewball/urlshortener/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
