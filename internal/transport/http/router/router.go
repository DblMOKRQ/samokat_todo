package router

import (
	"net/http"
	"time"

	"samokat_todo/internal/transport/http/middleware"
)

type TodoHandlers interface {
	CreateTodo(w http.ResponseWriter, r *http.Request)
	GetTodos(w http.ResponseWriter, r *http.Request)
	GetTodoByID(w http.ResponseWriter, r *http.Request)
	UpdateTodo(w http.ResponseWriter, r *http.Request)
	DeleteTodo(w http.ResponseWriter, r *http.Request)
}

func NewRouter(h TodoHandlers) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /todos", h.CreateTodo)
	mux.HandleFunc("GET /todos", h.GetTodos)

	mux.HandleFunc("GET /todos/{id}", h.GetTodoByID)
	mux.HandleFunc("PUT /todos/{id}", h.UpdateTodo)
	mux.HandleFunc("DELETE /todos/{id}", h.DeleteTodo)
	return middleware.Chain(
		mux,
		middleware.RequestID(),
		middleware.Recover(),
		middleware.Logger(),
	)
}

func NewHTTPServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
}
