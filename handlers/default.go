package handlers

import (
	"log/slog"
	"net/http"

	"github.com/RossHammer/go-todo/components"
)

type DefaultHandler struct {
	http.ServeMux
	log *slog.Logger
}

func New(log *slog.Logger) *DefaultHandler {
	h := &DefaultHandler{log: log}
	h.HandleFunc("GET /", wrapError(log, h.index))
	return h
}

func (h *DefaultHandler) index(w http.ResponseWriter, r *http.Request) error {
	v := components.Hello("Test")
	return v.Render(r.Context(), w)
}
