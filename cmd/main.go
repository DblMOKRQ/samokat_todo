package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"samokat_todo/internal/config"
	"samokat_todo/internal/repository/memory"
	"samokat_todo/internal/service"
	"samokat_todo/internal/transport/http/handlers"
	"samokat_todo/internal/transport/http/router"
	"syscall"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	repo := memory.NewTodoRepository()
	svc := service.NewTodoService(repo)

	h := handlers.NewHandler(svc, cfg.HTTP.RequestTimeout)

	handler := router.NewRouter(h)

	srv := router.NewHTTPServer(cfg.HTTP.Addr, handler)

	go func() {
		log.Printf("http server started on %s", cfg.HTTP.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("listen and serve error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Printf("shutdown signal received")

	ctx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown error: %v", err)
	} else {
		log.Printf("server stopped gracefully")
	}
}
