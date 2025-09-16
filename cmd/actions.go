package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"github.com/anewball/urlshortener/internal/jsonutil"
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

type ResultResponse struct {
	ShortCode string `json:"shortCode"`
	RawURL    string `json:"rawUrl"`
}

type DeleteResponse struct {
	Deleted   bool   `json:"deleted"`
	ShortCode string `json:"shortCode"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Details string `json:"details,omitempty"`
}

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
		return jsonutil.WriteJSON(out, ErrorResponse{Error: ErrLenZero.Error()})
	}

	arg := args[0]
	shortCode, err := svc.Add(ctx, arg)
	if err != nil {
		return jsonutil.WriteJSON(out, ErrorResponse{Error: fmt.Errorf("%w: %v", ErrAdd, err).Error()})
	}

	response := ResultResponse{ShortCode: shortCode, RawURL: arg}

	return jsonutil.WriteJSON(out, response)
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

	response := ResultResponse{ShortCode: arg, RawURL: url}

	return jsonutil.WriteJSON(out, response)
}

type ListResponse struct {
	Items  []ResultResponse `json:"items"`
	Count  int              `json:"count"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
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

	var results []ResultResponse = make([]ResultResponse, 0, len(urlItems))
	for _, u := range urlItems {
		results = append(results, ResultResponse{ShortCode: u.ShortCode, RawURL: u.OriginalURL})
	}

	response := ListResponse{
		Items:  results,
		Count:  len(results),
		Limit:  limit,
		Offset: offset,
	}

	return jsonutil.WriteJSON(out, response)
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

	return jsonutil.WriteJSON(out, response)
}
