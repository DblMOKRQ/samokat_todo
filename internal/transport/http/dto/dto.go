package dto

import (
	"time"

	"samokat_todo/internal/domain"
)

// То, что принимаем в POST/PUT
type TodoUpsertRequest struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}

// То, что отдаем наружу
type TodoResponse struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func ToTodoResponse(t domain.Todo) TodoResponse {
	return TodoResponse{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Completed:   t.Completed,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

func ToTodoResponses(list []domain.Todo) []TodoResponse {
	out := make([]TodoResponse, 0, len(list))
	for _, t := range list {
		out = append(out, ToTodoResponse(t))
	}
	return out
}
