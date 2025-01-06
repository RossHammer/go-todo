package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-h/templ"
	"github.com/google/uuid"
)

type ContextKey string

const (
	RequestID = ContextKey("requestId")
	Method    = ContextKey("method")
	Path      = ContextKey("path")
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func LogRequest(log *slog.Logger, inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r = r.WithContext(context.WithValue(r.Context(), RequestID, uuid.NewString()))
		r = r.WithContext(context.WithValue(r.Context(), Method, r.Method))
		r = r.WithContext(context.WithValue(r.Context(), Path, r.URL.Path))

		start := time.Now()
		lrw := &loggingResponseWriter{w, http.StatusOK}
		inner.ServeHTTP(lrw, r)
		duration := time.Since(start)
		log.InfoContext(r.Context(), "Completed request",
			slog.Int("status", lrw.statusCode),
			slog.Int64("durationMs", duration.Milliseconds()),
			slog.String("duration", duration.String()),
		)
	})
}

type ContextHandler struct {
	slog.Handler
}

func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if reqId, ok := ctx.Value(RequestID).(string); ok {
		r.AddAttrs(slog.String("requestId", reqId))
	}
	if method, ok := ctx.Value(Method).(string); ok {
		r.AddAttrs(slog.String("method", method))
	}
	if path, ok := ctx.Value(Path).(string); ok {
		r.AddAttrs(slog.String("path", path))
	}

	return h.Handler.Handle(ctx, r)
}

func wrapError(log *slog.Logger, f func(w http.ResponseWriter, r *http.Request) error) http.HandlerFunc {
	renderError := func(w http.ResponseWriter, r *http.Request, err error) {
		log.ErrorContext(r.Context(), "Error handling request", slog.Any("error", err))
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			renderError(w, r, err)
		}
	}
}

func wrapPage(log *slog.Logger, f func(w http.ResponseWriter, r *http.Request) (templ.Component, error)) http.HandlerFunc {
	return wrapError(log, func(w http.ResponseWriter, r *http.Request) error {
		c, err := f(w, r)
		if err != nil {
			return err
		}
		if err := c.Render(r.Context(), w); err != nil {
			return err
		}
		return nil
	})
}
