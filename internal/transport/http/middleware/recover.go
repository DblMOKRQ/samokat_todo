package middleware

import (
	"log"
	"net/http"
	"runtime/debug"
	"samokat_todo/internal/transport/http/constants"
)

func Recover() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					log.Printf("panic recovered: %v\n%s", rec, debug.Stack())
					http.Error(w, constants.MsgInternalError, http.StatusInternalServerError)
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
