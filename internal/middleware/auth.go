package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/kbannyi/shortener/internal/auth"
	"github.com/kbannyi/shortener/internal/logger"
)

const HEADER_NAME = "Authorization"

func RequireAuthMiddleware(h http.Handler) http.Handler {
	return process(h, false)
}

func AutoAuthMiddleware(h http.Handler) http.Handler {
	return process(h, true)
}

func process(h http.Handler, createUser bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Log.Infow("Auto auth for HTTP request:",
			"method", r.Method,
			"path", r.URL.Path,
		)

		headerValue := r.Header.Get(HEADER_NAME)
		var user auth.AuthUser
		var err error
		if headerValue == "" {
			if !createUser {
				http.Error(w, auth.ErrNotAuthenticated.Error(), http.StatusUnauthorized)
				return
			}
			user = auth.AuthUser{UserID: uuid.New().String()}
		} else {
			user, err = auth.ReadJWTString(headerValue)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		token, err := auth.BuildJWTString(user)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set(HEADER_NAME, token)

		ctx := auth.ToContext(r.Context(), user)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
