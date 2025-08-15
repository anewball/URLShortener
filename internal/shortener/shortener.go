package shortener

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DatabaseConn interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

type Shortener interface {
	Add(ctx context.Context, url string) (string, error)
	Get(ctx context.Context, shortCode string) (string, error)
	List(ctx context.Context, limit, offset int) ([]URLItem, error)
	Delete(ctx context.Context, shortCode string) error
}

type shortener struct {
	db DatabaseConn
}

type URLItem struct {
	ID          uint64
	OriginalURL string
	ShortCode   string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
}

func NewShortener(db DatabaseConn) *shortener {
	return &shortener{
		db: db,
	}
}

const (
	AddQuery    = "INSERT INTO url (original_url, short_code) VALUES ($1, $2) RETURNING id;"
	GetQuery    = "SELECT original_url FROM url WHERE short_code = $1"
	ListQuery   = "SELECT id, original_url, short_code, created_at, expires_at FROM url ORDER BY created_at DESC LIMIT $1 OFFSET $2"
	DeleteQuery = "DELETE FROM url WHERE short_code = $1"
	empty       = ""
)

func (s *shortener) Add(ctx context.Context, url string) (string, error) {
	if err := isValidURL(url); err != nil {
		return empty, fmt.Errorf("invalid URL: %w", err)
	}

	for range 3 {
		code, err := generateCode(7)
		if err != nil {
			return empty, err
		}

		_, err = s.db.Exec(ctx, AddQuery, url, code)
		if err == nil {
			return code, nil
		}

		var pqErr *pgconn.PgError
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			continue // Unique violation, try again
		}
		return empty, fmt.Errorf("insert: %w", err)
	}

	return empty, fmt.Errorf("failed to generate unique code after multiple attempts")
}

func (s *shortener) Get(ctx context.Context, shortCode string) (string, error) {
	if shortCode == empty {
		return empty, fmt.Errorf("short URL cannot be empty")
	}

	var originalURL string
	err := s.db.QueryRow(ctx, GetQuery, shortCode).Scan(&originalURL)
	if err != nil {
		if err == pgx.ErrNoRows {
			return empty, fmt.Errorf("short URL not found")
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

	var urlItems []URLItem
	for rows.Next() {
		var urlItem URLItem
		if err := rows.Scan(&urlItem.ID, &urlItem.OriginalURL, &urlItem.ShortCode, &urlItem.CreatedAt, &urlItem.ExpiresAt); err != nil {
			return nil, fmt.Errorf("failed to scan URL: %w", err)
		}
		urlItems = append(urlItems, urlItem)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}

	if len(urlItems) == 0 {
		return nil, fmt.Errorf("no URLs found")
	}

	return urlItems, nil
}

func (s *shortener) Delete(ctx context.Context, shortCode string) error {
	if shortCode == empty {
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
	if input == empty {
		return fmt.Errorf("URL cannot be empty")
	}
	s := strings.TrimSpace(input)
	u, err := url.Parse(s)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return errors.New("URL must include scheme (http/https) and host")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("only http/https are supported")
	}
	return nil
}

func generateCode(n int) (string, error) {
	alphabet := []byte("ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789")
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return empty, fmt.Errorf("random read: %w", err)
	}
	for i := range n {
		b[i] = alphabet[int(b[i])%len(alphabet)]
	}
	return string(b), nil
}
