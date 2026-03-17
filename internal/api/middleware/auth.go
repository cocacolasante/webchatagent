package middleware

import (
	"net/http"
)

// AdminAuthMiddleware validates the Blueprint admin API key for admin routes.
func AdminAuthMiddleware(adminKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := r.Header.Get("X-Blueprint-Admin-Key")
			if key == "" {
				key = r.URL.Query().Get("admin_key")
			}

			if key != adminKey || adminKey == "" {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
