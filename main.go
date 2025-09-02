package main

import (
	"fmt"
	"os"

	"github.com/anewball/urlshortener/cmd"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

func main() {
	godotenv.Load()
	viper.AutomaticEnv()

	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
