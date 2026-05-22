package domain

import "time"

// Todo представляет собой основную модель задачи
type Todo struct {
	ID          int
	Title       string
	Description string
	Completed   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
