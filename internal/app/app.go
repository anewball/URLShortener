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
	return &App{
		Conn:      conn,
		Shortener: shortener.New(conn),
	}
}

func (app *App) Close() {
	if app.Conn != nil {
		app.Conn.Close()
	}
}
