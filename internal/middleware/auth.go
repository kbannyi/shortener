package middleware

import (
	"net/http"

	"github.com/kbannyi/shortener/internal/logger"
)

func RequireAuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Infow("Require auth for HTTP request:",
			"method", r.Method,
			"path", r.URL.Path,
		)
		h.ServeHTTP(w, r)
	})
}

func AutoAuthMiddleware(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Infow("Auto auth for HTTP request:",
			"method", r.Method,
			"path", r.URL.Path,
		)
		h.ServeHTTP(w, r)
	})
}
