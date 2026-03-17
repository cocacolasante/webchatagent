package api

import (
	"net/http"

	"github.com/blueprintautomation/blueprint-chat/internal/api/handlers"
	mw "github.com/blueprintautomation/blueprint-chat/internal/api/middleware"
	"github.com/blueprintautomation/blueprint-chat/internal/tenant"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"
)

// RouterDeps holds all dependencies needed to build the router.
type RouterDeps struct {
	TenantService   *tenant.Service
	TenantHandler   *handlers.TenantsHandler
	ChatHandler     *handlers.ChatHandler
	LeadsHandler    *handlers.LeadsHandler
	BookingHandler  *handlers.BookingHandler
	WidgetHandler   *handlers.WidgetHandler
	HealthHandler   *handlers.HealthHandler
	RateLimiter     *mw.RateLimiter
	AdminKey        string
	RateLimitPerMin int
	Log             *zap.Logger
}

// NewRouter builds and returns the chi router with all routes configured.
func NewRouter(deps RouterDeps) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Recoverer)
	r.Use(mw.CORSMiddleware())
	r.Use(zapMiddleware(deps.Log))

	// Health check (no auth)
	r.Get("/api/health", deps.HealthHandler.Health)

	// Widget bundle (public)
	r.Get("/widget.js", deps.WidgetHandler.ServeWidget)

	// Widget config (public, requires tenant)
	r.With(mw.TenantMiddleware(deps.TenantService, deps.Log)).
		Get("/widget-config", deps.WidgetHandler.ServeConfig)

	// Public chat + lead + booking endpoints (require tenant)
	r.Group(func(r chi.Router) {
		r.Use(mw.TenantMiddleware(deps.TenantService, deps.Log))
		r.Use(deps.RateLimiter.TenantDailyLimitMiddleware())

		// Chat streaming
		r.With(deps.RateLimiter.ChatRateLimitMiddleware(deps.RateLimitPerMin)).
			Post("/api/chat/stream", deps.ChatHandler.Stream)

		// Lead capture
		r.Post("/api/leads", deps.LeadsHandler.Capture)

		// Booking
		r.Post("/api/booking/availability", deps.BookingHandler.GetAvailability)
		r.Post("/api/booking/create", deps.BookingHandler.CreateBooking)
	})

	// Admin endpoints (require admin key)
	r.Group(func(r chi.Router) {
		r.Use(mw.AdminAuthMiddleware(deps.AdminKey))
		r.Use(mw.PartnerScopeMiddleware)

		r.Route("/api/admin", func(r chi.Router) {
			// Tenant CRUD
			r.Post("/tenants", deps.TenantHandler.Create)
			r.Get("/tenants", deps.TenantHandler.List)
			r.Get("/tenants/{id}", deps.TenantHandler.Get)
			r.Put("/tenants/{id}", deps.TenantHandler.Update)
			r.Delete("/tenants/{id}", deps.TenantHandler.Delete)
			r.Post("/tenants/{id}/rotate-key", deps.TenantHandler.RotateKey)
			r.Post("/tenants/{id}/knowledge", deps.TenantHandler.UpdateKnowledge)

			// Tenant data
			r.Get("/tenants/{id}/leads", deps.LeadsHandler.List)

			// Blueprint Command stats
			r.Get("/tenants/{id}/stats", deps.TenantHandler.GetStats)
		})
	})

	return r
}

// zapMiddleware returns a chi-compatible logging middleware using zap.
func zapMiddleware(log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ww := chimw.NewWrapResponseWriter(w, r.ProtoMajor)
			next.ServeHTTP(ww, r)
			log.Info("request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", ww.Status()),
				zap.String("requestId", chimw.GetReqID(r.Context())),
			)
		})
	}
}
