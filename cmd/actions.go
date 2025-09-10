package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/anewball/urlshortener/internal/shortener"
)

type Actions interface {
	AddAction(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error
	GetAction(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error
	ListAction(ctx context.Context, limit int, offset int, out io.Writer, service shortener.URLShortener) error
	DeleteAction(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error
}

type actions struct{}

func NewActions() Actions {
	return &actions{}
}

func (a *actions) AddAction(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(args) == 0 {
		return errors.New("requires at least 1 arg(s), only received 0")
	}

	arg := args[0]
	code, err := service.Add(ctx, arg)
	if err != nil {
		return errors.New("failed to add URL")
	}

	result := Result{Code: code, Url: arg}

	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(result)
}

func (a *actions) GetAction(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(args) == 0 {
		return errors.New("requires at least 1 arg(s), only received 0")
	}

	arg := args[0]
	url, err := service.Get(ctx, arg)
	if err != nil {
		return errors.New("failed to retrieve original URL")
	}

	result := Result{Code: arg, Url: url}
	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(result)
}

func (a *actions) ListAction(ctx context.Context, limit int, offset int, out io.Writer, service shortener.URLShortener) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const maxLimit = 50
	if limit <= 0 || limit > maxLimit {
		return fmt.Errorf("limit must be between 1 and %d", maxLimit)
	}

	if offset < 0 {
		return fmt.Errorf("offset cannot be negative")
	}

	urlItems, err := service.List(ctx, limit, offset)
	if err != nil {
		return errors.New("failed to list URLs")
	}

	var results []Result = make([]Result, 0, len(urlItems))
	for _, u := range urlItems {
		results = append(results, Result{Code: u.ShortCode, Url: u.OriginalURL})
	}

	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(results)
}

func (a *actions) DeleteAction(ctx context.Context, out io.Writer, service shortener.URLShortener, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(args) == 0 {
		return fmt.Errorf("requires at least 1 arg(s), only received 0")
	}
	code := args[0]

	deleted, err := service.Delete(ctx, code)
	if err != nil {
		return errors.New("failed to delete URL")
	}
	if !deleted {
		return fmt.Errorf("no URL found for code %q", code)
	}

	var response DeleteResponse

	response.Deleted = deleted
	response.Code = code

	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(response)
}
