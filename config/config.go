package config

import (
	"errors"
	"strconv"
	"time"

	"github.com/anewball/urlshortener/env"
)

type Config struct {
	User            string
	Password        string
	Database        string
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

type Builder struct {
	db Config
	en env.Env
}

func NewBuilder(en env.Env) *Builder {
	return &Builder{db: Config{}, en: en}
}

func (b *Builder) MergeEnv() *Builder {
	if v, err := b.en.Get("DB_URL"); err == nil {
		b.db.URL = v
	}
	if v, err := b.en.Get("POSTGRES_USER"); err == nil {
		b.db.User = v
	}
	if v, err := b.en.Get("POSTGRES_PASSWORD"); err == nil {
		b.db.Password = v
	}
	if v, err := b.en.Get("POSTGRES_DB"); err == nil {
		b.db.Database = v
	}
	if v, err := b.en.Get("DB_MAX_CONNS"); err == nil {
		if n, err := strconv.Atoi(v); err == nil {
			b.db.MaxConns = int32(n)
		}
	}
	if v, err := b.en.Get("DB_MIN_CONNS"); err == nil {
		if n, err := strconv.Atoi(v); err == nil {
			b.db.MinConns = int32(n)
		}
	}
	if v, err := b.en.Get("DB_MAX_CONN_LIFETIME"); err == nil {
		if d, err := time.ParseDuration(v); err == nil {
			b.db.MaxConnLifetime = d
		}
	}
	if v, err := b.en.Get("DB_MAX_CONN_IDLE_TIME"); err == nil {
		if d, err := time.ParseDuration(v); err == nil {
			b.db.MaxConnIdleTime = d
		}
	}
	return b
}

func (b *Builder) validate() error {
	if b.db.URL == "" {
		return errors.New("URL is required")
	}
	if b.db.MaxConns < 0 {
		return errors.New("MaxConns must be >= 0")
	}
	if b.db.MinConns < 0 {
		return errors.New("MinConns must be >= 0")
	}
	if b.db.MaxConnLifetime < 0 {
		return errors.New("MaxConnLifetime must be >= 0")
	}
	if b.db.MaxConnIdleTime < 0 {
		return errors.New("MaxConnIdleTime must be >= 0")
	}
	if b.db.MaxConns > 0 && b.db.MinConns > b.db.MaxConns {
		return errors.New("MinConns must be <= MaxConns")
	}
	if b.db.Password == "" {
		return errors.New("password is required")
	}
	if b.db.User == "" {
		return errors.New("user is required")
	}
	if b.db.Database == "" {
		return errors.New("database is required")
	}
	return nil
}

func (b *Builder) Build() (Config, error) {
	if err := b.validate(); err != nil {
		return Config{}, err
	}
	return b.db, nil
}
