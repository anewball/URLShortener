package env

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Env interface {
	Get(key string) (string, error)
}

type env struct{}

func New() Env {
	return &env{}
}

func (*env) Get(key string) (string, error) {
	if !viper.IsSet(key) {
		return "", fmt.Errorf("key not found: %s", key)
	}
	v := viper.GetString(key)
	if strings.TrimSpace(v) == "" {
		return "", fmt.Errorf("key is empty: %s", key)
	}
	return v, nil
}
