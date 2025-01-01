package main

import (
	"embed"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/RossHammer/go-todo/handlers"
	"github.com/ory/graceful"
)

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
	log.Info("Starting server...")
	mux := http.NewServeMux()
	if dev {
		mux.Handle("GET /assets/", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	} else {
		mux.Handle("GET /assets/", http.FileServer(http.FS(assets)))
	}
	mux.Handle("/", handlers.New(log))

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
