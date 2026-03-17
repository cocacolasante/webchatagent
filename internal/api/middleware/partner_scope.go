package middleware

import (
	"context"
	"net/http"
	"strings"
)

type contextKey string

const (
	PartnerIDKey    contextKey = "partner_id"
	CallerTypeKey   contextKey = "caller_type"
	IsSuperAdminKey contextKey = "is_super_admin"
)

type CallerType string

const (
	CallerBPASuper CallerType = "bpa_super"
	CallerPartner  CallerType = "partner"
	CallerRegular  CallerType = "regular"
)

// PartnerScopeMiddleware injects partner scope into context based on API key prefix.
// Key format for partners: bpa_partner_{partnerUUID}_{random}
func PartnerScopeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := r.Header.Get("X-Blueprint-Admin-Key")
		if key == "" {
			key = r.URL.Query().Get("admin_key")
		}
		ctx := r.Context()

		switch {
		case strings.HasPrefix(key, "bpa_super_"):
			ctx = context.WithValue(ctx, CallerTypeKey, CallerBPASuper)
			ctx = context.WithValue(ctx, IsSuperAdminKey, true)

		case strings.HasPrefix(key, "bpa_partner_"):
			partnerID := parsePartnerID(key)
			if partnerID == "" {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"invalid partner key"}`, http.StatusUnauthorized)
				return
			}
			ctx = context.WithValue(ctx, CallerTypeKey, CallerPartner)
			ctx = context.WithValue(ctx, PartnerIDKey, partnerID)

		default:
			ctx = context.WithValue(ctx, CallerTypeKey, CallerRegular)
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetPartnerScope returns (partnerID, true) if caller is a partner, ("", false) otherwise.
func GetPartnerScope(ctx context.Context) (string, bool) {
	callerType, _ := ctx.Value(CallerTypeKey).(CallerType)
	if callerType != CallerPartner {
		return "", false
	}
	pid, _ := ctx.Value(PartnerIDKey).(string)
	return pid, pid != ""
}

func parsePartnerID(key string) string {
	remainder := strings.TrimPrefix(key, "bpa_partner_")
	if len(remainder) < 36 {
		return ""
	}
	candidate := remainder[:36]
	if isValidUUID(candidate) {
		return candidate
	}
	return ""
}

func isValidUUID(s string) bool {
	if len(s) != 36 {
		return false
	}
	for i, c := range s {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return false
			}
		} else if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
