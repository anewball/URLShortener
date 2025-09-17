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
	ErrLimit          = errors.New("the limit has to be at least 1 and cannot exceeds")
	ErrNegativeOffset = errors.New("offset cannot be negative")
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
		switch {
		case errors.Is(err, shortener.ErrIsValidURL):
			return jsonutil.WriteJSON(out, ErrorResponse{Error: "Invalid URL"})
		case errors.Is(err, shortener.ErrGenerate):
			return jsonutil.WriteJSON(out, ErrorResponse{Error: "Failed to produce short code"})
		case errors.Is(err, shortener.ErrQueryRow):
			return jsonutil.WriteJSON(out, ErrorResponse{Error: "Failed to add URL, please try again"})
		default:
			return jsonutil.WriteJSON(out, ErrorResponse{Error: "Failed to add URL"})
		}
	}

	response := ResultResponse{ShortCode: shortCode, RawURL: arg}

	return jsonutil.WriteJSON(out, response)
}

func (a *actions) GetAction(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(args) == 0 {
		return jsonutil.WriteJSON(out, ErrorResponse{Error: ErrLenZero.Error()})
	}

	arg := args[0]
	url, err := svc.Get(ctx, arg)
	if err != nil {
		switch {
		case errors.Is(err, shortener.ErrEmptyShortCode):
			return jsonutil.WriteJSON(out, ErrorResponse{Error: "A short code is required"})
		case errors.Is(err, shortener.ErrNotFound):
			return jsonutil.WriteJSON(out, ErrorResponse{Error: fmt.Sprintf("Could not find URL with short code %s", arg)})
		case errors.Is(err, shortener.ErrQuery):
			return jsonutil.WriteJSON(out, ErrorResponse{Error: "Failed to retrieve URL because of timeout"})
		default:
			return jsonutil.WriteJSON(out, ErrorResponse{Error: "Something went wrong"})
		}
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
		return jsonutil.WriteJSON(out, ErrorResponse{Error: fmt.Sprintf("%s: %d", ErrLimit.Error(), max)})
	}

	if offset < 0 {
		return jsonutil.WriteJSON(out, ErrorResponse{Error: fmt.Errorf("%w: %d", ErrNegativeOffset, offset).Error()})
	}

	urlItems, err := svc.List(ctx, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, shortener.ErrQuery):
			return jsonutil.WriteJSON(out, ErrorResponse{Error: fmt.Sprintf("Failed to retrieve URLs with limit: %d and offset: %d", limit, offset)})
		case errors.Is(err, shortener.ErrScan):
			return jsonutil.WriteJSON(out, ErrorResponse{Error: fmt.Sprintf("Failed to smarshal URLs with limit: %d and offset: %d", limit, offset)})
		case errors.Is(err, shortener.ErrRows):
			return jsonutil.WriteJSON(out, ErrorResponse{Error: fmt.Sprintf("An error occurs when smarshal URLs with limit: %d and offset: %d", limit, offset)})
		default:
			return jsonutil.WriteJSON(out, ErrorResponse{Error: fmt.Sprintf("An error occurs when retrieving URLs from limit: %d and offset: %d", limit, offset)})
		}
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
		return jsonutil.WriteJSON(out, ErrorResponse{Error: ErrLenZero.Error()})
	}
	shortCode := args[0]

	deleted, err := svc.Delete(ctx, shortCode)
	if err != nil {
		switch {
		case errors.Is(err, shortener.ErrEmptyShortCode):
			return jsonutil.WriteJSON(out, ErrorResponse{Error: "A short code is required"})
		case errors.Is(err, shortener.ErrExec):
			return jsonutil.WriteJSON(out, ErrorResponse{Error: fmt.Sprintf("A problem occurs when deleting short code: %s", shortCode)})
		case errors.Is(err, shortener.ErrNotFound):
			return jsonutil.WriteJSON(out, ErrorResponse{Error: fmt.Sprintf("Could not delete URL with short code %s", shortCode)})
		default:
			return jsonutil.WriteJSON(out, ErrorResponse{Error: fmt.Sprintf("Service could not delete URL with short code %s", shortCode)})
		}
	}
	if !deleted {
		return jsonutil.WriteJSON(out, ErrorResponse{Error: fmt.Sprintf("Problem deleting URL with short code %q", shortCode)})
	}

	var response DeleteResponse

	response.Deleted = deleted
	response.ShortCode = shortCode

	return jsonutil.WriteJSON(out, response)
}
