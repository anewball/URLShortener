package cmd

import "os"

type Env interface {
	Get(string) string
}

type realEnv struct {
}

func (realEnv) Get(k string) string {
	return os.Getenv(k)
}
