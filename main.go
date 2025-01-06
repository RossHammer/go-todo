package main

import (
	"database/sql"
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/RossHammer/go-todo/db"
	"github.com/RossHammer/go-todo/handlers"
	"github.com/ory/graceful"
	"github.com/pressly/goose/v3"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

//go:embed assets/*
var assets embed.FS

func main() {
	log := slog.New(&handlers.ContextHandler{Handler: slog.NewTextHandler(os.Stderr, nil)})

	flagSet := flag.NewFlagSet("server", flag.ExitOnError)
	dev := flagSet.Bool("dev", false, "run in development mode")
	if err := flagSet.Parse(os.Args[1:]); err != nil {
		log.Error("Error parsing flags", slog.Any("error", err))
		os.Exit(1)
	}

	if err := run(log, *dev); err != nil {
		log.Error("Error running server", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(log *slog.Logger, dev bool) error {
	db, err := sql.Open("sqlite", "file:todo.db")
	if err != nil {
		return fmt.Errorf("error opening database: %v", err)
	}
	defer db.Close()

	if err := runMigrations(log, db); err != nil {
		return fmt.Errorf("error running migrations: %v", err)
	}

	if err := runServer(log, dev, db); err != nil {
		return fmt.Errorf("error starting server: %v", err)
	}
	return nil
}

type gooseLogger struct{ *slog.Logger }

func (l gooseLogger) Fatalf(format string, a ...interface{}) {
	panic(fmt.Sprintf(format, a...))
}

func (l gooseLogger) Printf(format string, a ...interface{}) {
	l.Logger.Info(strings.TrimSpace(fmt.Sprintf(format, a...)))
}

func runMigrations(log *slog.Logger, db *sql.DB) error {
	log.Info("Running migrations...")
	goose.SetBaseFS(migrations)
	goose.SetLogger(gooseLogger{log})

	if err := goose.SetDialect("sqlite"); err != nil {
		return fmt.Errorf("error setting dialect: %v", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("error running migrations: %v", err)
	}

	log.Info("Migrations complete")
	return nil
}

func runServer(log *slog.Logger, dev bool, conn *sql.DB) error {
	log.Info("Starting server...")
	queries := db.New(conn)
	mux := http.NewServeMux()
	if dev {
		mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	} else {
		mux.Handle("GET /assets/", http.FileServer(http.FS(assets)))
	}
	mux.Handle("/", handlers.New(log, queries))

	server := graceful.WithDefaults(&http.Server{
		Addr:    ":8080",
		Handler: handlers.LogRequest(log, mux),
	})
	log.Info(fmt.Sprintf("Listening on %s", server.Addr))
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		return fmt.Errorf("error starting server: %v", err)
	}
	log.Info("Server stopped")
	return nil
}
