package middleware

import (
	"log"
	"net/http"
	"samokat_todo/internal/transport/http/constants"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(b []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(b)
	w.bytes += n
	return n, err
}

func Logger() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			sw := &statusWriter{ResponseWriter: w}

			next.ServeHTTP(sw, r)

			dur := time.Since(start)
			rid, _ := r.Context().Value(constants.RequestIDKey).(string)
			log.Printf("[%s] %s %s -> %d (%d bytes) in %s",
				rid, r.Method, r.URL.Path, sw.status, sw.bytes, dur,
			)
		})
	}
}
