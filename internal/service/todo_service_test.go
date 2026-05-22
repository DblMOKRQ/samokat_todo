package service

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"samokat_todo/internal/domain"
	"samokat_todo/internal/repository/memory"
)

type stubRepo struct {
	created   []*domain.Todo
	listResp  []domain.Todo
	byID      map[int]domain.Todo
	createErr error
}

func (s *stubRepo) Create(ctx context.Context, todo *domain.Todo) error {
	if s.createErr != nil {
		return s.createErr
	}
	todo.ID = len(s.created) + 1
	s.created = append(s.created, todo)
	return nil
}

func (s *stubRepo) List(ctx context.Context, limit, offset int) ([]domain.Todo, error) {
	return s.listResp, nil
}
func (s *stubRepo) GetByID(ctx context.Context, id int) (domain.Todo, error) {
	if t, ok := s.byID[id]; ok {
		return t, nil
	}
	return domain.Todo{}, domain.ErrTodoNotFound
}
func (s *stubRepo) Update(ctx context.Context, todo *domain.Todo) error { return nil }
func (s *stubRepo) Delete(ctx context.Context, id int) error            { return nil }

func TestTodoService_Create_Success(t *testing.T) {
	repo := &stubRepo{}
	svc := NewTodoService(repo)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	todo := &domain.Todo{
		Title:       "test",
		Description: "desc",
		Completed:   false,
	}

	if err := svc.Create(ctx, todo); err != nil {
		t.Fatalf("Create() error: %v", err)
	}
	if todo.ID == 0 {
		t.Fatalf("expected ID to be set")
	}
	if todo.CreatedAt.IsZero() {
		t.Fatalf("expected CreatedAt to be set")
	}
	if todo.UpdatedAt.IsZero() {
		t.Fatalf("expected UpdatedAt to be set")
	}
}

func TestTodoService_Create_ValidationError(t *testing.T) {
	repo := memory.NewTodoRepository()
	svc := NewTodoService(repo)

	ctx := context.Background()
	todo := &domain.Todo{Title: "   "}

	err := svc.Create(ctx, todo)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrInvalidTitle) {
		t.Fatalf("expected ErrInvalidTitle, got %v", err)
	}
}

func TestTodoService_Update_NotFound(t *testing.T) {
	repo := memory.NewTodoRepository()
	svc := NewTodoService(repo)

	ctx := context.Background()

	todo := &domain.Todo{
		ID:          999,
		Title:       "new title",
		Description: "d",
		Completed:   true,
	}

	err := svc.Update(ctx, todo)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrTodoNotFound) {
		t.Fatalf("expected ErrTodoNotFound, got %v", err)
	}
}

func TestTodoService_Update_ValidationError(t *testing.T) {
	repo := memory.NewTodoRepository()
	svc := NewTodoService(repo)

	ctx := context.Background()

	// Сначала создадим валидную задачу
	todo := &domain.Todo{Title: "ok"}
	if err := svc.Create(ctx, todo); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	// Пытаемся обновить пустым title
	upd := &domain.Todo{
		ID:          todo.ID,
		Title:       "   ",
		Description: "x",
		Completed:   false,
	}

	err := svc.Update(ctx, upd)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !errors.Is(err, domain.ErrInvalidTitle) {
		t.Fatalf("expected ErrInvalidTitle, got %v", err)
	}
}

func TestTodoService_Update_PreservesCreatedAtAndBumpsUpdatedAt(t *testing.T) {
	repo := memory.NewTodoRepository()
	svc := NewTodoService(repo)

	ctx := context.Background()

	todo := &domain.Todo{
		Title:       "t1",
		Description: "d1",
		Completed:   false,
	}
	if err := svc.Create(ctx, todo); err != nil {
		t.Fatalf("Create() error: %v", err)
	}

	createdAt := todo.CreatedAt
	updatedAt := todo.UpdatedAt

	time.Sleep(2 * time.Millisecond)

	upd := &domain.Todo{
		ID:          todo.ID,
		Title:       "t2",
		Description: "d2",
		Completed:   true,
	}
	if err := svc.Update(ctx, upd); err != nil {
		t.Fatalf("Update() error: %v", err)
	}

	if !upd.CreatedAt.Equal(createdAt) {
		t.Fatalf("expected CreatedAt to be preserved, got %v want %v", upd.CreatedAt, createdAt)
	}
	if !upd.UpdatedAt.After(updatedAt) {
		t.Fatalf("expected UpdatedAt to be bumped, got %v was %v", upd.UpdatedAt, updatedAt)
	}
}

func TestTodoRepository_ConcurrentCreate_NoDuplicateIDs(t *testing.T) {
	repo := memory.NewTodoRepository()
	ctx := context.Background()

	const goroutines = 100
	const iterations = 50
	var wg sync.WaitGroup

	idChan := make(chan int, goroutines*iterations)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				todo := &domain.Todo{
					Title: "Concurrent task",
				}
				if err := repo.Create(ctx, todo); err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				idChan <- todo.ID
			}
		}(i)
	}

	wg.Wait()
	close(idChan)

	// Проверяем все ID на уникальность
	seenIDs := make(map[int]bool)
	for id := range idChan {
		if seenIDs[id] {
			t.Errorf("found duplicate ID: %d", id)
		}
		seenIDs[id] = true
	}

	expectedTotal := goroutines * iterations
	if len(seenIDs) != expectedTotal {
		t.Errorf("expected %d unique IDs, but got %d", expectedTotal, len(seenIDs))
	}
}
