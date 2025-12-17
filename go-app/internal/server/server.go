package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"

	"go-app/internal/telemetry"
	"go-app/internal/todo"
)

// Server wires HTTP handlers.
type Server struct {
	repo   *todo.Repository
	logger telemetry.Logger
}

// New constructs a Server and returns a chi router.
func New(repo *todo.Repository, logger telemetry.Logger) *Server {
	return &Server{repo: repo, logger: logger}
}

// Router builds the HTTP router with all routes registered.
func (s *Server) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/health", s.handleHealth)

	r.Route("/api/v1/todos", func(r chi.Router) {
		r.Get("/", s.handleList)
		r.Post("/", s.handleCreate)
		r.Get("/{id}", s.handleGet)
		r.Put("/{id}", s.handleUpdate)
		r.Delete("/{id}", s.handleDelete)
	})

	return r
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "todo-api-go",
	})
}

func (s *Server) handleList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	skip, limit, err := parsePagination(r)
	if err != nil {
		s.logger.Error(ctx, "invalid pagination", zap.String("error", err.Error()))
		writeError(w, http.StatusBadRequest, err)
		return
	}

	todos, err := s.repo.List(ctx, skip, limit)
	if err != nil {
		s.logger.Error(ctx, "list todos failed", zap.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.logger.Info(ctx, "todos listed",
		zap.Int("count", len(todos)),
		zap.Int("skip", skip),
		zap.Int("limit", limit),
	)
	writeJSON(w, http.StatusOK, todos)
}

func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		s.logger.Error(ctx, "invalid todo id", zap.String("error", err.Error()))
		writeError(w, http.StatusBadRequest, err)
		return
	}

	todoItem, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, todo.ErrNotFound) {
			s.logger.Info(ctx, "todo not found", zap.String("todo.id", id.String()))
			writeError(w, http.StatusNotFound, fmt.Errorf("todo with id %s not found", id))
			return
		}
		s.logger.Error(ctx, "get todo failed",
			zap.String("todo.id", id.String()),
			zap.String("error", err.Error()),
		)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.logger.Info(ctx, "todo retrieved", zap.String("todo.id", todoItem.ID.String()))
	writeJSON(w, http.StatusOK, todoItem)
}

func (s *Server) handleCreate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var payload todo.CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.logger.Error(ctx, "invalid json", zap.String("error", err.Error()))
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	payload.Title = strings.TrimSpace(payload.Title)

	if err := validateCreate(payload); err != nil {
		s.logger.Error(ctx, "validation failed", zap.String("error", err.Error()))
		writeError(w, http.StatusBadRequest, err)
		return
	}

	todoItem, err := s.repo.Create(ctx, payload)
	if err != nil {
		s.logger.Error(ctx, "create todo failed", zap.String("error", err.Error()))
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.logger.Info(ctx, "todo created",
		zap.String("todo.id", todoItem.ID.String()),
		zap.String("title", todoItem.Title),
	)
	writeJSON(w, http.StatusCreated, todoItem)
}

func (s *Server) handleUpdate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var payload todo.UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		s.logger.Error(ctx, "invalid json", zap.String("error", err.Error()))
		writeError(w, http.StatusBadRequest, fmt.Errorf("invalid JSON: %w", err))
		return
	}

	if payload.Title != nil {
		trimmed := strings.TrimSpace(*payload.Title)
		payload.Title = &trimmed
	}

	if err := validateUpdate(payload); err != nil {
		s.logger.Error(ctx, "validation failed",
			zap.String("error", err.Error()),
			zap.String("todo.id", id.String()),
		)
		writeError(w, http.StatusBadRequest, err)
		return
	}

	todoItem, err := s.repo.Update(ctx, id, payload)
	if err != nil {
		if errors.Is(err, todo.ErrNotFound) {
			s.logger.Info(ctx, "todo not found for update", zap.String("todo.id", id.String()))
			writeError(w, http.StatusNotFound, fmt.Errorf("todo with id %s not found", id))
			return
		}
		if errors.Is(err, todo.ErrNoFieldsToUpdate) {
			s.logger.Error(ctx, "no fields to update", zap.String("todo.id", id.String()))
			writeError(w, http.StatusBadRequest, err)
			return
		}
		s.logger.Error(ctx, "update todo failed",
			zap.String("todo.id", id.String()),
			zap.String("error", err.Error()),
		)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.logger.Info(ctx, "todo updated", zap.String("todo.id", todoItem.ID.String()))
	writeJSON(w, http.StatusOK, todoItem)
}

func (s *Server) handleDelete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := parseUUIDParam(r, "id")
	if err != nil {
		s.logger.Error(ctx, "invalid todo id", zap.String("error", err.Error()))
		writeError(w, http.StatusBadRequest, err)
		return
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, todo.ErrNotFound) {
			s.logger.Info(ctx, "todo not found for delete", zap.String("todo.id", id.String()))
			writeError(w, http.StatusNotFound, fmt.Errorf("todo with id %s not found", id))
			return
		}
		s.logger.Error(ctx, "delete todo failed",
			zap.String("todo.id", id.String()),
			zap.String("error", err.Error()),
		)
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	s.logger.Info(ctx, "todo deleted", zap.String("todo.id", id.String()))
	w.WriteHeader(http.StatusNoContent)
}

// parsePagination extracts skip/limit query params with defaults.
func parsePagination(r *http.Request) (int, int, error) {
	// Defaults
	skip := 0
	limit := 100

	if v := r.URL.Query().Get("skip"); v != "" {
		val, err := strconv.Atoi(v)
		if err != nil || val < 0 {
			return 0, 0, fmt.Errorf("skip must be a non-negative integer")
		}
		skip = val
	}

	if v := r.URL.Query().Get("limit"); v != "" {
		val, err := strconv.Atoi(v)
		if err != nil || val < 1 || val > 500 {
			return 0, 0, fmt.Errorf("limit must be between 1 and 500")
		}
		limit = val
	}

	return skip, limit, nil
}

func parseUUIDParam(r *http.Request, key string) (uuid.UUID, error) {
	raw := chi.URLParam(r, key)
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s must be a valid UUID", key)
	}
	return id, nil
}

func validateCreate(payload todo.CreateRequest) error {
	if payload.Title == "" {
		return fmt.Errorf("title is required")
	}
	if len(payload.Title) > 200 {
		return fmt.Errorf("title must be at most 200 characters")
	}
	return nil
}

func validateUpdate(payload todo.UpdateRequest) error {
	if payload.Title == nil && payload.Description == nil && payload.Completed == nil {
		return todo.ErrNoFieldsToUpdate
	}
	if payload.Title != nil {
		if *payload.Title == "" {
			return fmt.Errorf("title, if provided, must not be empty")
		}
		if len(*payload.Title) > 200 {
			return fmt.Errorf("title must be at most 200 characters")
		}
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"detail": err.Error()})
}
