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
	ErrLenZero           = errors.New("requires at least 1 arg(s), only received 0")
	ErrLimit             = errors.New("limit must be between 1 and")
	ErrOffset            = errors.New("offset must be >= 0; got")
	ErrURLFormat         = errors.New("invalid URL format")
	ErrGenCode           = errors.New("could not create a short link at this time. Please try again later")
	ErrAdd               = errors.New("could not create a short link. Please try again")
	ErrUnsupported       = errors.New("error not supported")
	ErrShorCode          = errors.New("shortCode is required. Usage: get <shortCode>")
	ErrNotFound          = errors.New("no short link found for code")
	ErrTimeout           = errors.New("request timed out while retrieving the short link. Please try again later")
	ErrUnexpected        = errors.New("unexpected error. Please try again later")
	ErrQuery             = errors.New("an error occurred while retrieving URLs")
	ErrScan              = errors.New("unable to retrieve URLs. Please try again later")
	ErrRows              = errors.New("failed to marshal URLs")
	ErrUnknownList       = errors.New("failed to retrieve URLs")
	ErrDelete            = errors.New("unable to delete shortCode")
	ErrDeleteUnsupported = errors.New("service could not delete URL with short code")
	ErrUnableToDelete = errors.New("unable to delete short code")
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
		return writeAndReturnError(out, ErrLenZero, nil)
	}

	arg := args[0]
	shortCode, err := svc.Add(ctx, arg)
	if err != nil {
		switch {
		case errors.Is(err, shortener.ErrIsValidURL):
			return writeAndReturnError(out, ErrURLFormat, err)
		case errors.Is(err, shortener.ErrGenerate):
			return writeAndReturnError(out, ErrGenCode, err)
		case errors.Is(err, shortener.ErrQueryRow):
			return writeAndReturnError(out, ErrAdd, err)
		default:
			return writeAndReturnError(out, ErrUnsupported, err)
		}
	}

	response := ResultResponse{ShortCode: shortCode, RawURL: arg}

	return jsonutil.WriteJSON(out, response)
}

func (a *actions) GetAction(ctx context.Context, out io.Writer, svc shortener.URLShortener, args []string) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(args) == 0 {
		return writeAndReturnError(out, ErrLenZero, nil)
	}

	arg := args[0]
	url, err := svc.Get(ctx, arg)
	if err != nil {
		switch {
		case errors.Is(err, shortener.ErrEmptyShortCode):
			return writeAndReturnError(out, ErrShorCode, err)
		case errors.Is(err, shortener.ErrNotFound):
			return writeAndReturnError(out, fmt.Errorf("%w: %s", ErrNotFound, arg), err)
		case errors.Is(err, shortener.ErrQuery):
			return writeAndReturnError(out, ErrTimeout, err)
		default:
			return writeAndReturnError(out, ErrUnexpected, err)
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
		return writeAndReturnError(out, fmt.Errorf("%s: %d", ErrLimit.Error(), max), nil)
	}

	if offset < 0 {
		return writeAndReturnError(out, fmt.Errorf("%s: %d", ErrOffset.Error(), offset), nil)
	}

	urlItems, err := svc.List(ctx, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, shortener.ErrQuery):
			return writeAndReturnError(out, fmt.Errorf("%s (limit=%d, offset=%d)", ErrQuery, limit, offset), err)
		case errors.Is(err, shortener.ErrScan):
			return writeAndReturnError(out, ErrScan, err)
		case errors.Is(err, shortener.ErrRows):
			return writeAndReturnError(out, fmt.Errorf("%s (limit=%d, offset=%d)", ErrRows, limit, offset), err)
		default:
			return writeAndReturnError(out, fmt.Errorf("%s (limit=%d, offset=%d)", ErrRows, limit, offset), err)
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
		return writeAndReturnError(out, ErrLenZero, nil)
	}
	shortCode := args[0]

	deleted, err := svc.Delete(ctx, shortCode)
	if err != nil {
		switch {
		case errors.Is(err, shortener.ErrEmptyShortCode):
			return writeAndReturnError(out, ErrShorCode, err)
		case errors.Is(err, shortener.ErrExec):
			return writeAndReturnError(out, fmt.Errorf("%s %s", ErrDelete.Error(), shortCode), err)
		case errors.Is(err, shortener.ErrNotFound):
			return writeAndReturnError(out, fmt.Errorf("%w: %s", ErrNotFound, shortCode), err)
		default:
			return writeAndReturnError(out, fmt.Errorf("%w: %s", ErrDeleteUnsupported, shortCode), err)
		}
	}
	if !deleted {
		return writeAndReturnError(out, fmt.Errorf("%w: %s", ErrUnableToDelete, shortCode), nil)
	}

	var response DeleteResponse

	response.Deleted = deleted
	response.ShortCode = shortCode

	return jsonutil.WriteJSON(out, response)
}

func writeAndReturnError(out io.Writer, code error, cause error) error {
	_ = jsonutil.WriteJSON(out, ErrorResponse{
		Error: code.Error(),
		Details: func() string {
			if cause != nil {
				return cause.Error()
			}
			return ""
		}(),
	})
	if cause != nil {
		return fmt.Errorf("%w: %v", code, cause)
	}
	return code
}
