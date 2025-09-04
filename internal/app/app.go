package app

import (
	"github.com/anewball/urlshortener/internal/db"
	"github.com/anewball/urlshortener/internal/shortener"
)

type App struct {
	Conn      db.Conn
	Shortener shortener.Service
}

func New(conn db.Conn) *App {
	gen := shortener.NewNanoID(shortener.Alphabet)

	short, err := shortener.New(conn, gen)
	if err != nil {
		return nil
	}

	return &App{Conn: conn, Shortener: short}
}

func (app *App) Close() {
	if app.Conn != nil {
		app.Conn.Close()
	}
}
