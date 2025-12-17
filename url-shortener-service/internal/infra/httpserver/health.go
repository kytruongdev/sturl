package httpserver

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kytruongdev/sturl/url-shortener-service/internal/repository/redis"
)

// HealthStatus represents the health check response structure.
type HealthStatus struct {
	Status    string            `json:"status"`
	Timestamp time.Time         `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"
)

// ReadinessChecker defines the interface for checking service readiness.
type ReadinessChecker interface {
	CheckReadiness(ctx context.Context) error
}

// dbChecker implements ReadinessChecker for database connectivity.
type dbChecker struct {
	db *sql.DB
}

func (c *dbChecker) CheckReadiness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.db.PingContext(ctx)
}

// redisChecker implements ReadinessChecker for Redis connectivity.
type redisChecker struct {
	client redis.RedisClient
}

func (c *redisChecker) CheckReadiness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.client.Ping(ctx).Err()
}

// ReadinessConfig holds the dependencies needed for readiness checks.
type ReadinessConfig struct {
	DB    *sql.DB
	Redis redis.RedisClient
}

// checkReadiness performs readiness checks on all configured dependencies.
// It returns a HealthStatus with detailed check results.
func checkReadiness(cfg ReadinessConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		checks := make(map[string]string)
		allHealthy := true

		// Check database connectivity
		if cfg.DB != nil {
			dbChecker := &dbChecker{db: cfg.DB}
			if err := dbChecker.CheckReadiness(ctx); err != nil {
				checks["database"] = "unhealthy: " + err.Error()
				allHealthy = false
			} else {
				checks["database"] = "healthy"
			}
		}

		// Check Redis connectivity
		if cfg.Redis != nil {
			redisChecker := &redisChecker{client: cfg.Redis}
			if err := redisChecker.CheckReadiness(ctx); err != nil {
				checks["redis"] = "unhealthy: " + err.Error()
				allHealthy = false
			} else {
				checks["redis"] = "healthy"
			}
		}

		status := HealthStatus{
			Status:    StatusHealthy,
			Timestamp: time.Now(),
			Checks:    checks,
		}

		if !allHealthy {
			status.Status = StatusUnhealthy
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		json.NewEncoder(w).Encode(status)
	}
}
