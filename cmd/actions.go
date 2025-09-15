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

var (
	ErrLenZero        = errors.New("requires at least 1 arg(s), only received 0")
	ErrAdd            = errors.New("failed to add URL")
	ErrGet            = errors.New("failed to retrieve original URL")
	ErrDelete         = errors.New("failed to delete URL")
	ErrNotFound       = errors.New("no URL found for code")
	ErrLimit          = errors.New("invalid limit")
	ErrNegativeOffset = errors.New("offset cannot be negative")
	ErrList           = errors.New("failed to list URLs")
)

type Actions interface {
	AddAction(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error
	GetAction(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error
	ListAction(ctx context.Context, limit int, offset int, out io.Writer, svc shortener.URLShortener) error
	DeleteAction(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error
}

type actions struct {
	listMaxLimit int
}

func NewActions(listMaxLimit int) Actions {
	return &actions{listMaxLimit: listMaxLimit}
}

func (a *actions) AddAction(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(args) == 0 {
		return fmt.Errorf("%w", ErrLenZero)
	}

	arg := args[0]
	shortCode, err := svc.Add(ctx, arg)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAdd, err)
	}

	result := Result{ShortCode: shortCode, RawURL: arg}

	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(result)
}

func (a *actions) GetAction(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(args) == 0 {
		return fmt.Errorf("%w", ErrLenZero)
	}

	arg := args[0]
	url, err := svc.Get(ctx, arg)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrGet, err)
	}

	result := Result{ShortCode: arg, RawURL: url}
	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(result)
}

func (a *actions) ListAction(ctx context.Context, limit int, offset int, out io.Writer, svc shortener.URLShortener) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	const defaultMax = 500
	max := a.listMaxLimit
	if max <= 0 {
		max = defaultMax
	}
	if limit <= 0 || limit > max {
		return fmt.Errorf("%w: %d", ErrLimit, limit)
	}

	if offset < 0 {
		return fmt.Errorf("%w: %d", ErrNegativeOffset, offset)
	}

	urlItems, err := svc.List(ctx, limit, offset)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrList, err)
	}

	var results []Result = make([]Result, 0, len(urlItems))
	for _, u := range urlItems {
		results = append(results, Result{ShortCode: u.ShortCode, RawURL: u.OriginalURL})
	}

	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(results)
}

func (a *actions) DeleteAction(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(args) == 0 {
		return fmt.Errorf("%w", ErrLenZero)
	}
	shortCode := args[0]

	deleted, err := svc.Delete(ctx, shortCode)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDelete, err)
	}
	if !deleted {
		return fmt.Errorf("%w: %q", ErrNotFound, shortCode)
	}

	var response DeleteResponse

	response.Deleted = deleted
	response.ShortCode = shortCode

	encoder := json.NewEncoder(out)
	encoder.SetEscapeHTML(false)

	return encoder.Encode(response)
}
