package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/RossHammer/go-todo/handlers"
	"github.com/ory/graceful"
)

func main() {
	log := slog.New(&handlers.ContextHandler{Handler: slog.NewJSONHandler(os.Stderr, nil)})
	if err := run(log); err != nil {
		log.Error("Error running server", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(log *slog.Logger) error {
	log.Info("Starting server...")
	h := handlers.New(log)

	server := graceful.WithDefaults(&http.Server{
		Addr:    ":8000",
		Handler: handlers.LogRequest(log, h),
	})
	log.Info(fmt.Sprintf("Listening on %s", server.Addr))
	if err := graceful.Graceful(server.ListenAndServe, server.Shutdown); err != nil {
		return fmt.Errorf("error starting server: %v", err)
	}
	log.Info("Server stopped")
	return nil
}
