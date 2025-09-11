package env

import (
	"fmt"
)

var (
	ErrKeyNotFound  = fmt.Errorf("key not found")
	ErrValueIsEmpty = fmt.Errorf("value is empty")
)

type Env interface {
	Get(key string) (string, error)
}

type env struct {
	envMap map[string]string
}

func New(v map[string]string) Env {
	return &env{envMap: v}
}

func (e *env) Get(key string) (string, error) {
	if _, ok := e.envMap[key]; !ok {
		return "", fmt.Errorf("%w: %s", ErrKeyNotFound, key)
	}
	v := e.envMap[key]
	if v == "" {
		return "", fmt.Errorf("%w: %s", ErrValueIsEmpty, key)
	}
	return v, nil
}
