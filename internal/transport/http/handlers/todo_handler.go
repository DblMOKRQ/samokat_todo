package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"samokat_todo/internal/transport/http/constants"
	"samokat_todo/internal/transport/http/dto"
	"samokat_todo/internal/transport/http/response"
	"strconv"
	"time"

	"samokat_todo/internal/domain"
)

type todoService interface {
	Create(ctx context.Context, todo *domain.Todo) error
	List(ctx context.Context, limit, offset int) ([]domain.Todo, error)
	GetByID(ctx context.Context, id int) (domain.Todo, error)
	Update(ctx context.Context, todo *domain.Todo) error
	Delete(ctx context.Context, id int) error
}

type Handler struct {
	svc            todoService
	requestTimeout time.Duration
}

func NewHandler(svc todoService, requestTimeout time.Duration) *Handler {
	if requestTimeout <= 0 {
		requestTimeout = 3 * time.Second
	}
	return &Handler{
		svc:            svc,
		requestTimeout: requestTimeout,
	}
}

func (h *Handler) withTimeout(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), h.requestTimeout)
}

// POST /todos
func (h *Handler) CreateTodo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := h.withTimeout(r)
	defer cancel()

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.TodoUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var syntaxErr *json.SyntaxError
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			response.WriteError(w, http.StatusRequestEntityTooLarge, constants.MsgRequestBodyLarge)
			return
		} else if errors.As(err, &syntaxErr) || errors.Is(err, http.ErrBodyReadAfterClose) {
			response.WriteError(w, http.StatusBadRequest, constants.MsgInvalidJSONBody)
		} else {
			response.WriteError(w, http.StatusBadRequest, constants.MsgBadRequest)
		}
		return
	}

	t := domain.Todo{
		Title:       req.Title,
		Description: req.Description,
		Completed:   req.Completed,
	}

	if err := h.svc.Create(ctx, &t); err != nil {
		writeDomainError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusCreated, dto.ToTodoResponse(t))
}

// GET /todos
func (h *Handler) GetTodos(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := h.withTimeout(r)
	defer cancel()

	limit, offset, err := parseLimitOffset(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, "invalid limit/offset")
		return
	}

	list, err := h.svc.List(ctx, limit, offset)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.ToTodoResponses(list))
}

// GET /todos/{id}
func (h *Handler) GetTodoByID(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := h.withTimeout(r)
	defer cancel()

	id, err := todoIDFromPath(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, constants.MsgInvalidID)
		return
	}

	t, err := h.svc.GetByID(ctx, id)
	if err != nil {
		writeDomainError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.ToTodoResponse(t))
}

// PUT /todos/{id}
func (h *Handler) UpdateTodo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := h.withTimeout(r)
	defer cancel()

	id, err := todoIDFromPath(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, constants.MsgInvalidID)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	var req dto.TodoUpsertRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		var syntaxErr *json.SyntaxError
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			response.WriteError(w, http.StatusRequestEntityTooLarge, constants.MsgRequestBodyLarge)
			return
		} else if errors.As(err, &syntaxErr) || errors.Is(err, http.ErrBodyReadAfterClose) {
			response.WriteError(w, http.StatusBadRequest, constants.MsgInvalidJSONBody)
		} else {
			response.WriteError(w, http.StatusBadRequest, constants.MsgBadRequest)
		}
		return
	}

	t := domain.Todo{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Completed:   req.Completed,
	}

	if err := h.svc.Update(ctx, &t); err != nil {
		writeDomainError(w, err)
		return
	}

	response.WriteJSON(w, http.StatusOK, dto.ToTodoResponse(t))
}

// DELETE /todos/{id}
func (h *Handler) DeleteTodo(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := h.withTimeout(r)
	defer cancel()

	id, err := todoIDFromPath(r)
	if err != nil {
		response.WriteError(w, http.StatusBadRequest, constants.MsgInvalidID)
		return
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		writeDomainError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func todoIDFromPath(r *http.Request) (int, error) {
	raw := r.PathValue(constants.ID)
	if raw == "" {
		return 0, domain.ErrInvalidID
	}
	id, err := strconv.Atoi(raw)
	if err != nil || id <= 0 {
		return 0, domain.ErrInvalidID
	}
	return id, nil
}

func parseLimitOffset(r *http.Request) (limit, offset int, err error) {
	q := r.URL.Query()

	if v := q.Get("limit"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, 0, err
		}
		limit = n
	}
	if v := q.Get("offset"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, 0, err
		}
		offset = n
	}
	return limit, offset, nil
}

func writeDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrInvalidTitle):
		response.WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrTodoNotFound):
		response.WriteError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrInvalidID):
		response.WriteError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, domain.ErrInvalidPagination):
		response.WriteError(w, http.StatusBadRequest, err.Error())
	default:
		response.WriteError(w, http.StatusInternalServerError, constants.MsgInternalError)
	}
}
