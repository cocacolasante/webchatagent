package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// RateLimiter provides per-key rate limiting using Redis sliding window.
type RateLimiter struct {
	redis  *redis.Client
	log    *zap.Logger
}

// NewRateLimiter creates a new RateLimiter.
func NewRateLimiter(redisClient *redis.Client, log *zap.Logger) *RateLimiter {
	return &RateLimiter{redis: redisClient, log: log}
}

// ChatRateLimitMiddleware enforces per-session message rate limits.
func (rl *RateLimiter) ChatRateLimitMiddleware(messagesPerMin int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			sessionID := r.URL.Query().Get("sid")
			if sessionID == "" {
				next.ServeHTTP(w, r)
				return
			}

			key := fmt.Sprintf("blueprint-chat:ratelimit:session:%s:min", sessionID)
			allowed, err := rl.checkLimit(r.Context(), key, messagesPerMin, time.Minute)
			if err != nil {
				rl.log.Error("rate limit check failed", zap.Error(err))
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "60")
				http.Error(w, `{"error":"You're sending messages too quickly. Please wait a moment."}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// TenantDailyLimitMiddleware enforces per-tenant daily message limits.
func (rl *RateLimiter) TenantDailyLimitMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			t := TenantFromContext(r.Context())
			if t == nil {
				next.ServeHTTP(w, r)
				return
			}

			key := fmt.Sprintf("blueprint-chat:ratelimit:tenant:%s:day", t.ID)
			allowed, err := rl.checkLimit(r.Context(), key, t.MaxMessagesPerDay, 24*time.Hour)
			if err != nil {
				rl.log.Error("tenant daily limit check failed", zap.Error(err))
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("Retry-After", "3600")
				http.Error(w, `{"error":"Daily message limit reached. Please try again tomorrow."}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// checkLimit uses a Redis counter with expiry for rate limiting.
func (rl *RateLimiter) checkLimit(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	pipe := rl.redis.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)

	if _, err := pipe.Exec(ctx); err != nil {
		return true, fmt.Errorf("redis pipeline: %w", err)
	}

	return incr.Val() <= int64(limit), nil
}
