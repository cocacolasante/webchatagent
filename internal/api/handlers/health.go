package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

// HealthHandler handles health check requests.
type HealthHandler struct {
	db    *pgxpool.Pool
	redis *redis.Client
}

// NewHealthHandler creates a new HealthHandler.
func NewHealthHandler(db *pgxpool.Pool, redis *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, redis: redis}
}

// Health returns the system health status.
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	dbStatus := "ok"
	if err := h.db.Ping(ctx); err != nil {
		dbStatus = "error: " + err.Error()
	}

	redisStatus := "ok"
	if err := h.redis.Ping(ctx).Err(); err != nil {
		redisStatus = "error: " + err.Error()
	}

	status := "ok"
	httpStatus := http.StatusOK
	if dbStatus != "ok" || redisStatus != "ok" {
		status = "degraded"
		httpStatus = http.StatusServiceUnavailable
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpStatus)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    status,
		"timestamp": time.Now().UTC(),
		"services": map[string]string{
			"postgres": dbStatus,
			"redis":    redisStatus,
		},
	})
}
