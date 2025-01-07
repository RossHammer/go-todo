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
	DeleteTodo(context.Context, int64) error
	UpdateTodo(context.Context, db.UpdateTodoParams) (db.Todo, error)
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
	h.HandleFunc("POST /add", wrapPage(log, h.add))
	h.HandleFunc("POST /update/{id}", wrapPage(log, h.check))
	h.HandleFunc("DELETE /delete/{id}", wrapError(log, h.delete))
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

func (h *DefaultHandler) add(w http.ResponseWriter, r *http.Request) (templ.Component, error) {
	var newTodo newTodo
	validation, err := h.formReader.ReadForm(&newTodo, r)
	if err != nil {
		return nil, fmt.Errorf("error reading form: %w", err)
	} else if validation != nil {
		_, err = w.Write([]byte("Validation failed: " + validation[0].Message))
		return nil, err
	}
	item, err := h.repository.AddTodo(r.Context(), newTodo.Title)
	if err != nil {
		return nil, fmt.Errorf("error adding todo: %w", err)
	}
	return components.TodoItem(item), nil
}

func (h *DefaultHandler) delete(w http.ResponseWriter, r *http.Request) error {
	id, err := util.ParseInt(r.PathValue("id"))
	if err != nil {
		return fmt.Errorf("error parsing id: %w", err)
	}

	if err := h.repository.DeleteTodo(r.Context(), id); err != nil {
		return fmt.Errorf("error deleting todo: %w", err)
	}
	return nil
}

type updateTodo struct {
	Completed *bool `validate:"required"`
}

func (h *DefaultHandler) check(w http.ResponseWriter, r *http.Request) (templ.Component, error) {
	var updateTodo updateTodo
	validation, err := h.formReader.ReadForm(&updateTodo, r)
	if err != nil {
		return nil, fmt.Errorf("error reading form: %w", err)
	} else if validation != nil {
		_, err = w.Write([]byte("Validation failed: " + validation[0].Message))
		return nil, err
	}

	id, err := util.ParseInt(r.PathValue("id"))
	if err != nil {
		return nil, fmt.Errorf("error parsing id: %w", err)
	}

	item, err := h.repository.UpdateTodo(r.Context(), db.UpdateTodoParams{ID: id, Completed: *updateTodo.Completed})
	if err != nil {
		return nil, fmt.Errorf("error updating todo: %w", err)
	}
	return components.TodoItem(item), nil
}
