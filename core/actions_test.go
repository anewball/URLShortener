package core

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/anewball/urlshortener/internal/jsonutil"
	"github.com/anewball/urlshortener/internal/shortener"
	"github.com/stretchr/testify/assert"
)

func TestAddActions(t *testing.T) {
	listMaxLimit := 20
	shortCode := "Hpa3t2B"
	testCases := []struct {
		name                   string
		args                   []string
		buf                    bytes.Buffer
		isError                bool
		expectedErrorResponse  ErrorResponse
		expectedResultResponse ResultResponse
		action                 Actions
		svc                    shortener.URLShortener
	}{
		{
			name:                   "success",
			args:                   []string{"https://example.com"},
			action:                 NewActions(listMaxLimit),
			buf:                    bytes.Buffer{},
			expectedResultResponse: ResultResponse{ShortCode: shortCode, RawURL: "https://example.com"},
			isError:                false,
			expectedErrorResponse:  ErrorResponse{},
			svc: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return shortCode, nil
				},
			},
		},
		{
			name:                  "zero args",
			args:                  []string{},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrLenZero.Error()},
			svc:                   &mockedShortener{},
		},
		{
			name:                  "invalid url",
			args:                  []string{"https://example.com"},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrURLFormat.Error(), Details: shortener.ErrIsValidURL.Error()},
			svc: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return "", shortener.ErrIsValidURL
				},
			},
		},
		{
			name:    "error empty args",
			args:    []string{"https://example.com"},
			action:  NewActions(listMaxLimit),
			buf:     bytes.Buffer{},
			isError: true,
			expectedErrorResponse: ErrorResponse{
				Error:   ErrAdd.Error(),
				Details: errors.New("error generating short code").Error(),
			},
			svc: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return "", shortener.ErrGenerate
				},
			},
		},
		{
			name:                  "could not add url",
			args:                  []string{"https://example.com"},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrAdd.Error(), Details: shortener.ErrQueryRow.Error()},
			svc: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return "", shortener.ErrQueryRow
				},
			},
		},
		{
			name:                  "error not supported",
			args:                  []string{"https://example.com"},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrUnsupported.Error(), Details: "Failed to add URL"},
			svc: &mockedShortener{
				addFunc: func(ctx context.Context, url string) (string, error) {
					return "", errors.New("Failed to add URL")
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := tc.action.AddAction(ctx, &tc.buf, tc.svc, tc.args)

			if tc.isError {
				var actualErrorResponse ErrorResponse
				jsonutil.ReadJSON(&tc.buf, &actualErrorResponse)
				assert.Equal(t, tc.expectedErrorResponse, actualErrorResponse)
				return
			}

			assert.NoError(t, err)

			var actualResultResponse ResultResponse
			jsonutil.ReadJSON(&tc.buf, &actualResultResponse)

			assert.Equal(t, tc.expectedResultResponse, actualResultResponse)
		})
	}
}

func TestGetAction(t *testing.T) {
	listMaxLimit := 20
	shortCode := "Hpa3t2B"
	testCases := []struct {
		name                   string
		args                   []string
		buf                    bytes.Buffer
		isError                bool
		expectedErrorResponse  ErrorResponse
		expectedResultResponse ResultResponse
		action                 Actions
		svc                    shortener.URLShortener
	}{
		{
			name:                   "success",
			args:                   []string{shortCode},
			action:                 NewActions(listMaxLimit),
			buf:                    bytes.Buffer{},
			expectedResultResponse: ResultResponse{ShortCode: shortCode, RawURL: "https://example.com"},
			isError:                false,
			expectedErrorResponse:  ErrorResponse{},
			svc: &mockedShortener{
				getFunc: func(ctx context.Context, shortCode string) (string, error) {
					return "https://example.com", nil
				},
			},
		},
		{
			name:                  "zero args",
			args:                  []string{},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrLenZero.Error()},
			svc:                   &mockedShortener{},
		},
		{
			name:    "error empty short code",
			args:    []string{""},
			action:  NewActions(20),
			buf:     bytes.Buffer{},
			isError: true,
			expectedErrorResponse: ErrorResponse{
				Error:   ErrShortCode.Error(),
				Details: errors.New("a required short code was not provided. Please see usage: get <shortCode>").Error(),
			},
			svc: &mockedShortener{
				getFunc: func(ctx context.Context, url string) (string, error) {
					return "", shortener.ErrShortCode
				},
			},
		},
		{
			name:                  "error not found",
			args:                  []string{shortCode},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Sprintf("%s: %s", ErrNotFound, shortCode), Details: shortener.ErrNotFound.Error()},
			svc: &mockedShortener{
				getFunc: func(ctx context.Context, url string) (string, error) {
					return "", shortener.ErrNotFound
				},
			},
		},
		{
			name:    "error query",
			args:    []string{shortCode},
			action:  NewActions(listMaxLimit),
			buf:     bytes.Buffer{},
			isError: true,
			expectedErrorResponse: ErrorResponse{
				Error:   ErrUnexpected.Error(),
				Details: errors.New("an error occurred while retrieving the short link. Please try again later").Error(),
			},
			svc: &mockedShortener{
				getFunc: func(ctx context.Context, url string) (string, error) {
					return "", shortener.ErrQuery
				},
			},
		},
		{
			name:                  "error",
			args:                  []string{shortCode},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrUnexpected.Error(), Details: "Something went wrong"},
			svc: &mockedShortener{
				getFunc: func(ctx context.Context, url string) (string, error) {
					return "", errors.New("Something went wrong")
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := tc.action.GetAction(ctx, &tc.buf, tc.svc, tc.args)

			if tc.isError {
				var actualErrorResponse ErrorResponse
				jsonutil.ReadJSON(&tc.buf, &actualErrorResponse)
				assert.Equal(t, tc.expectedErrorResponse, actualErrorResponse)
				return
			}

			assert.NoError(t, err)

			var actualResultResponse ResultResponse
			jsonutil.ReadJSON(&tc.buf, &actualResultResponse)

			assert.Equal(t, tc.expectedResultResponse, actualResultResponse)
		})
	}
}

func TestDeleteAction(t *testing.T) {
	listMaxLimit := 20
	shortCode := "Hpa3t2B"
	testCases := []struct {
		name                   string
		args                   []string
		buf                    bytes.Buffer
		isError                bool
		expectedErrorResponse  ErrorResponse
		expectedDeleteResponse DeleteResponse
		action                 Actions
		svc                    shortener.URLShortener
	}{
		{
			name:                   "success",
			args:                   []string{shortCode},
			action:                 NewActions(listMaxLimit),
			buf:                    bytes.Buffer{},
			expectedDeleteResponse: DeleteResponse{Deleted: true, ShortCode: shortCode},
			isError:                false,
			expectedErrorResponse:  ErrorResponse{},
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, shortCode string) (bool, error) {
					return true, nil
				},
			},
		},
		{
			name:                  "zero args",
			args:                  []string{},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: ErrLenZero.Error()},
			svc:                   &mockedShortener{},
		},
		{
			name:    "error empty short code",
			args:    []string{""},
			action:  NewActions(listMaxLimit),
			buf:     bytes.Buffer{},
			isError: true,
			expectedErrorResponse: ErrorResponse{
				Error:   ErrShortCode.Error(),
				Details: errors.New("a required short code was not provided. Please see usage: delete <shortCode>").Error(),
			},
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, shortener.ErrShortCode
				},
			},
		},
		{
			name:                  "error with exec",
			args:                  []string{shortCode},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Errorf("%s %s", ErrDelete.Error(), shortCode).Error(), Details: shortener.ErrExec.Error()},
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, shortener.ErrExec
				},
			},
		},
		{
			name:                  "error URL not found",
			args:                  []string{shortCode},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Errorf("%w: %s", ErrNotFound, shortCode).Error(), Details: shortener.ErrNotFound.Error()},
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, shortener.ErrNotFound
				},
			},
		},
		{
			name:    "unknown error",
			args:    []string{shortCode},
			action:  NewActions(listMaxLimit),
			buf:     bytes.Buffer{},
			isError: true,
			expectedErrorResponse: ErrorResponse{
				Error:   ErrUnexpected.Error(),
				Details: fmt.Errorf("failed to delete short code: %q", shortCode).Error(),
			},
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, errors.New("unknown error")
				},
			},
		},
		{
			name:                  "when deleted variable is false",
			args:                  []string{shortCode},
			action:                NewActions(listMaxLimit),
			buf:                   bytes.Buffer{},
			isError:               true,
			expectedErrorResponse: ErrorResponse{Error: fmt.Errorf("%w: %s", ErrUnableToDelete, shortCode).Error()},
			svc: &mockedShortener{
				deleteFunc: func(ctx context.Context, url string) (bool, error) {
					return false, nil
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := tc.action.DeleteAction(ctx, &tc.buf, tc.svc, tc.args)

			if tc.isError {
				var actualErrorResponse ErrorResponse
				jsonutil.ReadJSON(&tc.buf, &actualErrorResponse)
				assert.Equal(t, tc.expectedErrorResponse, actualErrorResponse)
				return
			}

			assert.NoError(t, err)

			var actualDeleteResponse DeleteResponse
			jsonutil.ReadJSON(&tc.buf, &actualDeleteResponse)

			assert.Equal(t, tc.expectedDeleteResponse, actualDeleteResponse)
		})
	}
}

func TestListAction(t *testing.T) {
	listMaxLimit := 20
	testCases := []struct {
		name                  string
		offset                int
		limit                 int
		buf                   bytes.Buffer
		isError               bool
		expectedErrorResponse ErrorResponse
		expectedListResponse  ListResponse
		action                Actions
		svc                   shortener.URLShortener
	}{
		{
			name:   "success",
			offset: 0,
			limit:  2,
			action: NewActions(listMaxLimit),
			buf:    bytes.Buffer{},
			expectedListResponse: ListResponse{
				Items: []ResultResponse{
					{RawURL: "https://anewball.com", ShortCode: "nMHdgTh"},
					{RawURL: "https://jayden.newball.com", ShortCode: "k5aBWD5"},
				}, Count: 2, Limit: 2, Offset: 0,
			},
			isError:               false,
			expectedErrorResponse: ErrorResponse{},
			svc: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{
						{ID: 1, OriginalURL: "https://anewball.com", ShortCode: "nMHdgTh", CreatedAt: time.Date(2025, time.August, 25, 14, 30, 0, 0, time.UTC), ExpiresAt: nil},
						{ID: 2, OriginalURL: "https://jayden.newball.com", ShortCode: "k5aBWD5", CreatedAt: time.Date(2025, time.August, 25, 14, 3, 0, 0, time.UTC), ExpiresAt: nil},
					}, nil
				},
			},
		},
		{
			name:   "success when limit max is zero",
			offset: 0,
			limit:  2,
			action: NewActions(0),
			buf:    bytes.Buffer{},
			expectedListResponse: ListResponse{
				Items: []ResultResponse{
					{RawURL: "https://anewball.com", ShortCode: "nMHdgTh"},
					{RawURL: "https://jayden.newball.com", ShortCode: "k5aBWD5"},
				}, Count: 2, Limit: 2, Offset: 0,
			},
			isError:               false,
			expectedErrorResponse: ErrorResponse{},
			svc: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{
						{ID: 1, OriginalURL: "https://anewball.com", ShortCode: "nMHdgTh", CreatedAt: time.Date(2025, time.August, 25, 14, 30, 0, 0, time.UTC), ExpiresAt: nil},
						{ID: 2, OriginalURL: "https://jayden.newball.com", ShortCode: "k5aBWD5", CreatedAt: time.Date(2025, time.August, 25, 14, 3, 0, 0, time.UTC), ExpiresAt: nil},
					}, nil
				},
			},
		},
		{
			name:                 "limit less than zero",
			offset:               0,
			limit:                -2,
			action:               NewActions(listMaxLimit),
			buf:                  bytes.Buffer{},
			expectedListResponse: ListResponse{},
			isError:              true,
			expectedErrorResponse: ErrorResponse{
				Error:   ErrLimit.Error(),
				Details: fmt.Errorf("limit must be between 1 and %d; got %d", listMaxLimit, -2).Error(),
			},
			svc: &mockedShortener{},
		},
		{
			name:                 "offset less than zero",
			offset:               -2,
			limit:                2,
			action:               NewActions(listMaxLimit),
			buf:                  bytes.Buffer{},
			expectedListResponse: ListResponse{},
			isError:              true,
			expectedErrorResponse: ErrorResponse{
				Error:   ErrOffset.Error(),
				Details: fmt.Errorf("offset must be >= 0; got %d", -2).Error(),
			},
			svc: &mockedShortener{},
		},
		{
			name:                 "error query",
			offset:               0,
			limit:                2,
			action:               NewActions(listMaxLimit),
			buf:                  bytes.Buffer{},
			expectedListResponse: ListResponse{},
			isError:              true,
			expectedErrorResponse: ErrorResponse{
				Error:   ErrUnexpected.Error(),
				Details: fmt.Errorf("error executing list query (limit=%d, offset=%d)", 2, 0).Error(),
			},
			svc: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{}, shortener.ErrQuery
				},
			},
		},
		{
			name:                 "error scan",
			offset:               0,
			limit:                2,
			action:               NewActions(listMaxLimit),
			buf:                  bytes.Buffer{},
			expectedListResponse: ListResponse{},
			isError:              true,
			expectedErrorResponse: ErrorResponse{
				Error:   ErrUnexpected.Error(),
				Details: fmt.Errorf("error scanning rows (limit=%d, offset=%d)", 2, 0).Error(),
			},
			svc: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{}, shortener.ErrScan
				},
			},
		},
		{
			name:                 "error rows",
			offset:               0,
			limit:                2,
			action:               NewActions(listMaxLimit),
			buf:                  bytes.Buffer{},
			expectedListResponse: ListResponse{},
			isError:              true,
			expectedErrorResponse: ErrorResponse{
				Error:   ErrUnexpected.Error(),
				Details: fmt.Errorf("row iteration error (limit=%d, offset=%d)", 2, 0).Error(),
			},
			svc: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{}, shortener.ErrRows
				},
			},
		},
		{
			name:                 "unknown error",
			offset:               0,
			limit:                2,
			action:               NewActions(listMaxLimit),
			buf:                  bytes.Buffer{},
			expectedListResponse: ListResponse{},
			isError:              true,
			expectedErrorResponse: ErrorResponse{
				Error:   ErrUnexpected.Error(),
				Details: fmt.Errorf("unknown list error (limit=%d, offset=%d)", 2, 0).Error(),
			},
			svc: &mockedShortener{
				listFunc: func(ctx context.Context, limit int, offset int) ([]shortener.URLItem, error) {
					return []shortener.URLItem{}, errors.New("something went wrong")
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			err := tc.action.ListAction(ctx, tc.limit, tc.offset, &tc.buf, tc.svc)

			if tc.isError {
				var actualErrorResponse ErrorResponse
				jsonutil.ReadJSON(&tc.buf, &actualErrorResponse)
				assert.Equal(t, tc.expectedErrorResponse, actualErrorResponse)
				return
			}

			assert.NoError(t, err)

			var actualListResponse ListResponse
			jsonutil.ReadJSON(&tc.buf, &actualListResponse)

			assert.Equal(t, tc.expectedListResponse, actualListResponse)
		})
	}
}
