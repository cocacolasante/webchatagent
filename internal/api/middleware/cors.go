package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORSMiddleware returns a CORS middleware that allows all origins.
// This is required because the widget embeds on any third-party website.
func CORSMiddleware() func(http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Blueprint-Admin-Key", "X-Tenant-ID"},
		ExposedHeaders:   []string{"Content-Type", "X-Request-ID"},
		AllowCredentials: false,
		MaxAge:           300,
	})
}
