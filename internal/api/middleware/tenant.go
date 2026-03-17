package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/blueprintautomation/blueprint-chat/internal/tenant"
	"go.uber.org/zap"
)

// TenantKey is the context key for the resolved tenant.
type TenantKey struct{}

// TenantMiddleware resolves the tenant from the request and injects it into context.
func TenantMiddleware(svc *tenant.Service, log *zap.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tenantID := resolveTenantID(r)
			if tenantID == "" {
				http.Error(w, `{"error":"tenant ID required"}`, http.StatusBadRequest)
				return
			}

			t, err := svc.GetByID(r.Context(), tenantID)
			if err != nil {
				log.Debug("tenant not found", zap.String("id", tenantID), zap.Error(err))
				http.Error(w, `{"error":"tenant not found"}`, http.StatusNotFound)
				return
			}

			if !t.IsActive {
				http.Error(w, `{"error":"tenant is inactive"}`, http.StatusForbidden)
				return
			}

			ctx := context.WithValue(r.Context(), TenantKey{}, t)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// TenantFromContext retrieves the tenant from a request context.
func TenantFromContext(ctx context.Context) *tenant.Tenant {
	t, _ := ctx.Value(TenantKey{}).(*tenant.Tenant)
	return t
}

// resolveTenantID attempts to find a tenant ID from multiple sources.
func resolveTenantID(r *http.Request) string {
	// 1. Query parameter
	if tid := r.URL.Query().Get("tid"); tid != "" {
		return tid
	}

	// 2. Header
	if tid := r.Header.Get("X-Tenant-ID"); tid != "" {
		return tid
	}

	// 3. Subdomain (future feature stub)
	host := r.Host
	if idx := strings.Index(host, "."); idx > 0 {
		sub := host[:idx]
		// Exclude "www", "api", "chat" — those aren't tenant subdomains
		if sub != "www" && sub != "api" && sub != "chat" && sub != "localhost" {
			return sub
		}
	}

	return ""
}
