package todo

import (
	"time"

	"github.com/google/uuid"
)

// Todo represents a task in the todos table.
type Todo struct {
	ID          uuid.UUID `db:"id" json:"id"`
	Title       string    `db:"title" json:"title"`
	Description *string   `db:"description" json:"description"`
	Completed   bool      `db:"completed" json:"completed"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
