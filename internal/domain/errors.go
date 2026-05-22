package domain

import "errors"

var (
	ErrTodoNotFound      = errors.New("todo not found")
	ErrInvalidTitle      = errors.New("title cannot be empty")
	ErrInvalidID         = errors.New("invalid id")
	ErrInvalidPagination = errors.New("invalid pagination")
)
