package main

import (
	"fmt"
	"os"

	"github.com/anewball/urlshortener/cmd"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file if needed
	// This can be done using a package like godotenv
	_ = godotenv.Load()

	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
