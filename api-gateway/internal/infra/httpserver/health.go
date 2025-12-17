package httpserver

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
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

// upstreamChecker implements ReadinessChecker for upstream service connectivity.
type upstreamChecker struct {
	serviceName string
	baseURL     string
	client      *http.Client
}

func (c *upstreamChecker) CheckReadiness(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	// Check upstream service health by hitting its health endpoint
	healthURL := c.baseURL + "/health/ready"
	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Consider 2xx and 3xx status codes as healthy
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return nil
	}

	return &UpstreamHealthError{
		ServiceName: c.serviceName,
		StatusCode:  resp.StatusCode,
		URL:         healthURL,
	}
}

// UpstreamHealthError represents an error when checking upstream service health.
type UpstreamHealthError struct {
	ServiceName string
	StatusCode  int
	URL         string
}

func (e *UpstreamHealthError) Error() string {
	return "upstream service unhealthy"
}

// UpstreamServiceConfig represents configuration for an upstream service health check.
type UpstreamServiceConfig struct {
	Name    string // Service name identifier
	BaseURL string // Base URL of the upstream service
}

// ReadinessConfig holds the dependencies needed for readiness checks.
type ReadinessConfig struct {
	UpstreamServices []UpstreamServiceConfig
}

// checkReadiness performs readiness checks on all configured upstream services.
// It returns a HealthStatus with detailed check results.
func checkReadiness(cfg ReadinessConfig) http.HandlerFunc {
	if cfg.UpstreamServices == nil || len(cfg.UpstreamServices) <= 0 {
		return nil
	}

	// Create HTTP client for health checks with short timeout
	client := &http.Client{
		Timeout: 3 * time.Second,
		Transport: &http.Transport{
			ResponseHeaderTimeout: 2 * time.Second,
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		checks := make(map[string]string)
		allHealthy := true

		// Check each upstream service
		for _, upstream := range cfg.UpstreamServices {
			checker := &upstreamChecker{
				serviceName: upstream.Name,
				baseURL:     upstream.BaseURL,
				client:      client,
			}

			if err := checker.CheckReadiness(ctx); err != nil {
				checks[upstream.Name] = "unhealthy: " + err.Error()
				allHealthy = false
			} else {
				checks[upstream.Name] = "healthy"
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
