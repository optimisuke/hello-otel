package todo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// Sentinel errors for common cases.
var (
	ErrNoFieldsToUpdate = errors.New("no fields to update")
	ErrNotFound         = errors.New("todo not found")
)

// Repository handles DB operations for todos.
type Repository struct {
	db *sqlx.DB
}

// NewRepository constructs a todo repository.
func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

// List returns todos ordered by creation date desc.
func (r *Repository) List(ctx context.Context, skip, limit int) ([]Todo, error) {
	todos := []Todo{}
	query := `
        SELECT id, title, description, completed, created_at, updated_at
        FROM todos
        ORDER BY created_at DESC
        OFFSET $1 LIMIT $2`
	if err := r.db.SelectContext(ctx, &todos, query, skip, limit); err != nil {
		return nil, err
	}
	return todos, nil
}

// Get fetches a todo by ID.
func (r *Repository) Get(ctx context.Context, id uuid.UUID) (*Todo, error) {
	var todo Todo
	query := `
        SELECT id, title, description, completed, created_at, updated_at
        FROM todos
        WHERE id = $1`
	if err := r.db.GetContext(ctx, &todo, query, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &todo, nil
}

// Create inserts a new todo.
func (r *Repository) Create(ctx context.Context, payload CreateRequest) (*Todo, error) {
	id := uuid.New()
	completed := false
	if payload.Completed != nil {
		completed = *payload.Completed
	}

	query := `
        INSERT INTO todos (id, title, description, completed)
        VALUES ($1, $2, $3, $4)
        RETURNING id, title, description, completed, created_at, updated_at`

	var todo Todo
	if err := r.db.GetContext(ctx, &todo, query, id, payload.Title, payload.Description, completed); err != nil {
		return nil, err
	}
	return &todo, nil
}

// Update modifies fields on an existing todo. At least one field must be non-nil.
func (r *Repository) Update(ctx context.Context, id uuid.UUID, payload UpdateRequest) (*Todo, error) {
	setParts := []string{}
	args := []any{}

	if payload.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", len(args)+1))
		args = append(args, *payload.Title)
	}
	if payload.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", len(args)+1))
		args = append(args, payload.Description)
	}
	if payload.Completed != nil {
		setParts = append(setParts, fmt.Sprintf("completed = $%d", len(args)+1))
		args = append(args, *payload.Completed)
	}

	if len(setParts) == 0 {
		return nil, ErrNoFieldsToUpdate
	}

	// Always bump updated_at server-side.
	setParts = append(setParts, fmt.Sprintf("updated_at = NOW()"))

	query := fmt.Sprintf(`
        UPDATE todos
        SET %s
        WHERE id = $%d
        RETURNING id, title, description, completed, created_at, updated_at`,
		strings.Join(setParts, ", "), len(args)+1,
	)

	args = append(args, id)

	var todo Todo
	if err := r.db.GetContext(ctx, &todo, query, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &todo, nil
}

// Delete removes a todo by ID.
func (r *Repository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM todos WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}
