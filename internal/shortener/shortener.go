package shortener

import (
	"context"
	"errors"
	"fmt"
	"net/url"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jxskiss/base62"
)

type DatabaseConn interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type Shortener interface {
	Add(ctx context.Context, url string) (string, error)
	Get(ctx context.Context, shortCode string) (string, error)
	List(ctx context.Context, limit, offset int) ([]string, error)
	Delete(ctx context.Context, shortCode string) error
}

type shortener struct {
	db       DatabaseConn
	alphabet string
	base     int
	store    map[string]string
}

func NewShortener(db DatabaseConn) *shortener {
	alphabet := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	base := len(alphabet)
	return &shortener{
		alphabet: alphabet,
		base:     base,
		store:    make(map[string]string),
		db:       db,
	}
}

const (
	AddQuery    = "INSERT INTO url (short_code, original_url) VALUES ($1, $2)"
	GetQuery    = "SELECT original_url FROM url WHERE short_code = $1"
	ListQuery   = "SELECT short_code FROM url ORDER BY created_at DESC LIMIT $1 OFFSET $2"
	DeleteQuery = "DELETE FROM url WHERE short_code = $1"
)

func (s *shortener) Add(ctx context.Context, url string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("URL cannot be empty")
	}

	if err := isValidURL(url); err != nil {
		return "", fmt.Errorf("invalid URL: %w", err)
	}

	u := uuid.New()
	encoded := base62.EncodeToString(u[:])

	_, err := s.db.Exec(ctx, AddQuery, encoded, url)
	if err != nil {
		return "", fmt.Errorf("failed to insert URL: %w", err)
	}

	return encoded, nil
}

func (s *shortener) Get(ctx context.Context, shortCode string) (string, error) {
	if shortCode == "" {
		return "", fmt.Errorf("short URL cannot be empty")
	}

	var originalURL string
	err := s.db.QueryRow(ctx, GetQuery, shortCode).Scan(&originalURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return "", fmt.Errorf("short URL not found")
		}
		return "", fmt.Errorf("failed to retrieve URL: %w", err)
	}

	return originalURL, nil
}

func (s *shortener) List(ctx context.Context, limit, offset int) ([]string, error) {
	rows, err := s.db.Query(ctx, ListQuery, limit, offset)
	if err != nil {
		return []string{}, fmt.Errorf("failed to list URLs: %w", err)
	}
	defer rows.Close()

	var urls []string
	for rows.Next() {
		var url string
		if err := rows.Scan(&url); err != nil {
			return []string{}, fmt.Errorf("failed to scan URL: %w", err)
		}
		urls = append(urls, url)
	}

	if err := rows.Err(); err != nil {
		return []string{}, fmt.Errorf("error iterating over rows: %w", err)
	}

	if len(urls) == 0 {
		return []string{}, fmt.Errorf("no URLs found")
	}

	return urls, nil
}

func (s *shortener) Delete(ctx context.Context, shortCode string) error {
	if shortCode == "" {
		return fmt.Errorf("short URL cannot be empty")
	}

	_, err := s.db.Exec(ctx, DeleteQuery, shortCode)
	if err != nil {
		if err == pgx.ErrNoRows {
			return fmt.Errorf("short URL not found")
		}

		return fmt.Errorf("failed to delete URL: %w", err)
	}

	return nil
}

func isValidURL(input string) error {
	parsedURL, err := url.Parse(input)
	if err != nil {
		return errors.New("invalid URL format")
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return errors.New("URL must have a scheme(http/https) and a host")
	}

	return nil
}
