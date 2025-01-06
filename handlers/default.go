package handlers

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/RossHammer/go-todo/components"
	"github.com/RossHammer/go-todo/db"
	"github.com/RossHammer/go-todo/util"
	"github.com/a-h/templ"
)

type TodoRepository interface {
	ListTodos(context.Context) ([]db.Todo, error)
	AddTodo(context.Context, string) (db.Todo, error)
}

type DefaultHandler struct {
	http.ServeMux
	log        *slog.Logger
	repository TodoRepository
	formReader *util.FormReader
}

func New(log *slog.Logger, repository TodoRepository) *DefaultHandler {
	h := &DefaultHandler{log: log, repository: repository, formReader: util.NewFormReader()}
	h.HandleFunc("GET /", wrapPage(log, h.index))
	h.HandleFunc("POST /add", wrapError(log, h.add))
	return h
}

func (h *DefaultHandler) index(w http.ResponseWriter, r *http.Request) (templ.Component, error) {
	todos, err := h.repository.ListTodos(r.Context())
	if err != nil {
		return nil, fmt.Errorf("error listing todos: %w", err)
	}

	return components.TodoPage(todos), nil
}

type newTodo struct {
	Title string `validate:"required,max=255"`
}

func (h *DefaultHandler) add(w http.ResponseWriter, r *http.Request) error {
	var newTodo newTodo
	validation, err := h.formReader.ReadForm(&newTodo, r)
	if err != nil {
		return fmt.Errorf("error reading form: %w", err)
	} else if validation != nil {
		_, err = w.Write([]byte("Validation failed: " + validation[0].Message))
		return err
	}

	_, err = h.repository.AddTodo(r.Context(), newTodo.Title)
	if err != nil {
		return fmt.Errorf("error adding todo: %w", err)
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
	return nil
}
