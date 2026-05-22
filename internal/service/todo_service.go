package service

import (
	"context"
	"strings"
	"time"

	"samokat_todo/internal/domain"
)

// TODO: Что с ними делать? Типа в конфиг?
const (
	DefaultLimit = 50
	MaxLimit     = 100
)

// TodoService реализует бизнес-логику работы с задачами
type TodoService struct {
	repo TodoRepository
}

type TodoRepository interface {
	Create(ctx context.Context, todo *domain.Todo) error
	List(ctx context.Context, limit, offset int) ([]domain.Todo, error)
	GetByID(ctx context.Context, id int) (domain.Todo, error)
	Update(ctx context.Context, todo *domain.Todo) error
	Delete(ctx context.Context, id int) error
}

// NewTodoService создает новый сервис с переданным репозиторием
func NewTodoService(repo TodoRepository) *TodoService {
	return &TodoService{
		repo: repo,
	}
}

// Create проверяет данные и передает их в репозиторий для сохранения
func (s *TodoService) Create(ctx context.Context, input *domain.Todo) error {
	input.Title = strings.TrimSpace(input.Title)
	if input.Title == "" {
		return domain.ErrInvalidTitle
	}

	now := time.Now()
	input.CreatedAt = now
	input.UpdatedAt = now

	return s.repo.Create(ctx, input)
}

func (s *TodoService) List(ctx context.Context, limit, offset int) ([]domain.Todo, error) {
	if offset < 0 {
		return nil, domain.ErrInvalidPagination
	}

	if limit == 0 {
		limit = DefaultLimit
	}
	if limit < 0 {
		return nil, domain.ErrInvalidPagination
	}
	if limit > MaxLimit {
		limit = MaxLimit
	}

	return s.repo.List(ctx, limit, offset)
}

func (s *TodoService) GetByID(ctx context.Context, id int) (domain.Todo, error) {
	if id <= 0 {
		return domain.Todo{}, domain.ErrInvalidID
	}
	return s.repo.GetByID(ctx, id)
}

func (s *TodoService) Update(ctx context.Context, todo *domain.Todo) error {
	todo.Title = strings.TrimSpace(todo.Title)
	if todo.Title == "" {
		return domain.ErrInvalidTitle
	}

	existing, err := s.repo.GetByID(ctx, todo.ID)
	if err != nil {
		return err
	}

	todo.CreatedAt = existing.CreatedAt
	todo.UpdatedAt = time.Now()

	return s.repo.Update(ctx, todo)
}

func (s *TodoService) Delete(ctx context.Context, id int) error {
	if id <= 0 {
		return domain.ErrInvalidID
	}
	return s.repo.Delete(ctx, id)
}
