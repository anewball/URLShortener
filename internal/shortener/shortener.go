package shortener

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/anewball/urlshortener/internal/db"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Shortener interface {
	Add(ctx context.Context, url string) (string, error)
	Get(ctx context.Context, shortCode string) (string, error)
	List(ctx context.Context, limit, offset int) ([]URLItem, error)
	Delete(ctx context.Context, shortCode string) (bool, error)
}

type shortener struct {
	db db.Conn
}

type URLItem struct {
	ID          uint64
	OriginalURL string
	ShortCode   string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
}

func New(db db.Conn) *shortener {
	return &shortener{
		db: db,
	}
}

const (
	AddQuery    = "INSERT INTO url (original_url, short_code) VALUES ($1, $2);"
	GetQuery    = "SELECT original_url FROM url WHERE short_code = $1 AND (expires_at IS NULL OR expires_at > now());"
	ListQuery   = "SELECT id, original_url, short_code, created_at, expires_at FROM url ORDER BY created_at DESC LIMIT $1 OFFSET $2;"
	DeleteQuery = "DELETE FROM url WHERE short_code = $1;"
	empty       = ""
	codeLen     = 7
	maxRetries  = 5
)

func (s *shortener) Add(ctx context.Context, url string) (string, error) {
	if err := isValidURL(url); err != nil {
		return empty, fmt.Errorf("invalid URL: %w", err)
	}

	for range maxRetries {
		code := generateCode(codeLen)

		_, err := s.db.Exec(ctx, AddQuery, url, code)
		if err == nil {
			return code, nil
		}

		if isUniqueViolation(err) {
			continue
		}

		return empty, err
	}

	return empty, errors.New("exhausted retries")
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
		return []URLItem{}, nil
	}

	return urlItems, nil
}

func (s *shortener) Delete(ctx context.Context, shortCode string) (bool, error) {
	if shortCode == empty {
		return false, fmt.Errorf("short URL cannot be empty")
	}

	cmdTag, err := s.db.Exec(ctx, DeleteQuery, shortCode)
	if err != nil {
		return false, err
	}

	return cmdTag.RowsAffected() > 0, nil
}

func isValidURL(raw string) error {
	if raw == empty {
		return fmt.Errorf("URL cannot be empty")
	}
	s := strings.TrimSpace(raw)
	u, err := url.Parse(s)
	if err != nil || u.Scheme == empty || u.Host == empty {
		return errors.New("URL must include scheme (http/https) and host")
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return errors.New("only http/https are supported")
	}
	return nil
}

func generateCode(n int) string {
	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	k := len(alphabet) // work in int
	out := make([]byte, n)
	var buf [1]byte

	for i := range n {
		for {
			if _, err := rand.Read(buf[:]); err != nil {
				panic("crypto/rand failed: " + err.Error())
			}
			x := int(buf[0])             // 0..255
			threshold := 256 - (256 % k) // in int, 0..256
			if x < threshold {           // accept only values < threshold
				out[i] = alphabet[x%k]
				break
			}
			// otherwise, draw again (rejection sampling)
		}
	}
	return string(out)
}

func isUniqueViolation(err error) bool {
	var pqErr *pgconn.PgError
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return true
	}
	return false
}
