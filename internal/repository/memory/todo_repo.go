package memory

import (
	"context"
	"samokat_todo/internal/domain"
	"sync"
)

// TodoRepository реализация in-memory хранилища
type TodoRepository struct {
	mu     sync.RWMutex
	todos  map[int]domain.Todo
	order  []int
	nextID int
}

// NewTodoRepository создает новый экземпляр репозитория
func NewTodoRepository() *TodoRepository {
	return &TodoRepository{
		todos:  make(map[int]domain.Todo),
		order:  make([]int, 0, 16),
		nextID: 1,
	}
}

// Create добавляет новую задачу
func (r *TodoRepository) Create(ctx context.Context, todo *domain.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	id := r.nextID
	todo.ID = id

	r.todos[id] = *todo
	r.order = append(r.order, id)
	r.nextID++
	return nil
}

// List возвращает список всех задач с пагинацией
func (r *TodoRepository) List(ctx context.Context, limit, offset int) ([]domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if offset > len(r.order) {
		return []domain.Todo{}, nil
	}
	end := len(r.order)
	if limit >= 0 && offset+limit < end {
		end = offset + limit
	}

	result := make([]domain.Todo, 0, end-offset)
	for _, id := range r.order[offset:end] {
		result = append(result, r.todos[id])
	}
	return result, nil
}

// GetByID ищет задачу по ID
func (r *TodoRepository) GetByID(ctx context.Context, id int) (domain.Todo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	todo, exists := r.todos[id]
	if !exists {
		return domain.Todo{}, domain.ErrTodoNotFound
	}

	return todo, nil
}

// Update обновляет существующую задачу
func (r *TodoRepository) Update(ctx context.Context, todo *domain.Todo) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, exists := r.todos[todo.ID]
	if !exists {
		return domain.ErrTodoNotFound
	}

	r.todos[todo.ID] = *todo
	return nil
}

// Delete удаляет задачу
func (r *TodoRepository) Delete(ctx context.Context, id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.todos[id]; !exists {
		return domain.ErrTodoNotFound
	}

	delete(r.todos, id)
	for i, elem := range r.order {
		if elem == id {
			r.order = append(r.order[:i], r.order[i+1:]...)
			break
		}
	}
	return nil
}
