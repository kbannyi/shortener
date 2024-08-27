package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/kbannyi/shortener/internal/auth"
	"github.com/kbannyi/shortener/internal/logger"
)

const HeaderName = "Authorization"

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

		headerValue := r.Header.Get(HeaderName)
		if headerValue == "" {
			c, err := r.Cookie(HeaderName)
			if err == nil {
				headerValue = c.Value
			}
		}
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
			logger.Log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}

		token, err := auth.BuildJWTString(user)
		if err != nil {
			logger.Log.Error(err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set(HeaderName, token)
		http.SetCookie(w, &http.Cookie{Name: HeaderName, Value: token})

		ctx := auth.ToContext(r.Context(), user)
		h.ServeHTTP(w, r.WithContext(ctx))
	})
}
