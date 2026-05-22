package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"samokat_todo/internal/transport/http/constants"
)

func RequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rid := r.Header.Get(constants.HeaderRequestID)
			if rid == "" {
				rid = newRequestID()
			}
			w.Header().Set(constants.HeaderRequestID, rid)
			ctx := context.WithValue(r.Context(), constants.RequestIDKey, rid)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
