package todo

// CreateRequest holds payload for creating a todo.
type CreateRequest struct {
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}

// UpdateRequest holds payload for updating a todo.
type UpdateRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}
