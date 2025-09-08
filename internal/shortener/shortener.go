package shortener

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/anewball/urlshortener/internal/db"
	"github.com/jackc/pgx/v5"
)

const maxURLLength = 2048

var (
	ErrNotFound  = errors.New("short URL not found")
	ErrEmptyCode = errors.New("short URL cannot be empty")
)

type URLShortener interface {
	Add(ctx context.Context, url string) (string, error)
	Get(ctx context.Context, shortCode string) (string, error)
	List(ctx context.Context, limit, offset int) ([]URLItem, error)
	Delete(ctx context.Context, shortCode string) (bool, error)
}

var _ URLShortener = (*shortener)(nil)

type shortener struct {
	db  db.Conn
	gen NanoID
}

type URLItem struct {
	ID          uint64
	OriginalURL string
	ShortCode   string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
}

func New(db db.Conn, gen NanoID) (URLShortener, error) {
	if db == nil {
		return nil, fmt.Errorf("db cannot be nil")
	}
	if gen == nil {
		gen = NewNanoID(Alphabet)
	}
	return &shortener{db: db, gen: gen}, nil
}

const (
	AddQuery    = "INSERT INTO url (original_url, short_code) VALUES ($1, $2);"
	GetQuery    = "SELECT original_url FROM url WHERE short_code = $1 AND (expires_at IS NULL OR expires_at > now());"
	ListQuery   = "SELECT id, original_url, short_code, created_at, expires_at FROM url ORDER BY created_at DESC LIMIT $1 OFFSET $2;"
	DeleteQuery = "DELETE FROM url WHERE short_code = $1;"
	empty       = ""
	codeLen     = 7
	Alphabet    = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
)

func (s *shortener) Add(ctx context.Context, rawURL string) (string, error) {
	if err := isValidURL(rawURL); err != nil {
		return empty, fmt.Errorf("invalid URL: %w", err)
	}

	id, err := s.gen.Generate(codeLen)
	if err != nil {
		return empty, err
	}

	_, err = s.db.Exec(ctx, AddQuery, rawURL, id)
	if err == nil {
		return id, nil
	}

	return empty, err
}

func (s *shortener) Get(ctx context.Context, shortCode string) (string, error) {
	if shortCode == empty {
		return empty, ErrEmptyCode
	}

	var originalURL string
	err := s.db.QueryRow(ctx, GetQuery, shortCode).Scan(&originalURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return empty, ErrNotFound
		}
		return empty, fmt.Errorf("failed to retrieve URL: %w", err)
	}

	return originalURL, nil
}

func (s *shortener) List(ctx context.Context, limit, offset int) ([]URLItem, error) {
	rows, err := s.db.Query(ctx, ListQuery, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list URLs: %w", err)
	}
	defer rows.Close()

	items := make([]URLItem, 0, limit)
	for rows.Next() {
		var item URLItem
		if err := rows.Scan(&item.ID, &item.OriginalURL, &item.ShortCode, &item.CreatedAt, &item.ExpiresAt); err != nil {
			return nil, fmt.Errorf("failed to scan URL: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	return items, nil
}

func (s *shortener) Delete(ctx context.Context, shortCode string) (bool, error) {
	if shortCode == empty {
		return false, ErrEmptyCode
	}

	cmdTag, err := s.db.Exec(ctx, DeleteQuery, shortCode)
	if err != nil {
		return false, fmt.Errorf("failed to delete: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return false, ErrNotFound
	}

	return true, nil
}

func isValidURL(rawURL string) error {
	if rawURL == empty {
		return ErrEmptyCode
	}

	s := strings.TrimSpace(rawURL)
	u, err := url.Parse(s)
	if err != nil || u.Scheme == empty || u.Host == empty {
		return errors.New("URL must include scheme (http/https) and host")
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("only http/https are supported")
	}

	if len(s) > maxURLLength {
		return fmt.Errorf("URL too long")
	}
	
	return nil
}
