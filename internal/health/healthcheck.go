package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opiagile/direito-lux/internal/auth"
	"github.com/opiagile/direito-lux/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusUnhealthy Status = "unhealthy"
	StatusDegraded  Status = "degraded"
)

// Check represents a health check result
type Check struct {
	Name      string                 `json:"name"`
	Status    Status                 `json:"status"`
	Message   string                 `json:"message,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration_ms"`
	Details   map[string]interface{} `json:"details,omitempty"`
}

// Response represents the overall health response
type Response struct {
	Status     Status                 `json:"status"`
	Version    string                 `json:"version"`
	Timestamp  time.Time              `json:"timestamp"`
	Checks     []Check                `json:"checks"`
	TotalTime  time.Duration          `json:"total_duration_ms"`
	SystemInfo map[string]interface{} `json:"system_info"`
}

// Checker interface for health checks
type Checker interface {
	Check(ctx context.Context) Check
}

// Handler manages health checks
type Handler struct {
	version       string
	checkers      map[string]Checker
	mu            sync.RWMutex
	cache         *redis.Client
	cacheDuration time.Duration
}

// NewHandler creates a new health check handler
func NewHandler(version string, redisClient *redis.Client) *Handler {
	return &Handler{
		version:       version,
		checkers:      make(map[string]Checker),
		cache:         redisClient,
		cacheDuration: 5 * time.Second,
	}
}

// Register adds a new health checker
func (h *Handler) Register(name string, checker Checker) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.checkers[name] = checker
}

// HealthCheckHandler returns the Gin handler for health checks
func (h *Handler) HealthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		verbose := c.Query("verbose") == "true"

		// Check cache first for non-verbose requests
		if !verbose && h.cache != nil {
			if cached, err := h.getFromCache(ctx); err == nil {
				c.JSON(http.StatusOK, cached)
				return
			}
		}

		response := h.performHealthChecks(ctx, verbose)

		// Cache the response
		if h.cache != nil && response.Status == StatusHealthy {
			h.cacheResponse(ctx, response)
		}

		// Set appropriate status code
		statusCode := http.StatusOK
		if response.Status == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		} else if response.Status == StatusDegraded {
			statusCode = http.StatusMultiStatus
		}

		c.JSON(statusCode, response)
	}
}

// LivenessHandler returns a simple liveness check
func (h *Handler) LivenessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "alive",
			"timestamp": time.Now(),
		})
	}
}

// ReadinessHandler returns readiness status
func (h *Handler) ReadinessHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Quick readiness checks
		checks := []Check{}
		overallStatus := StatusHealthy

		// Check critical components only
		criticalCheckers := []string{"database", "keycloak", "redis"}

		h.mu.RLock()
		for _, name := range criticalCheckers {
			if checker, exists := h.checkers[name]; exists {
				check := checker.Check(ctx)
				checks = append(checks, check)

				if check.Status == StatusUnhealthy {
					overallStatus = StatusUnhealthy
				}
			}
		}
		h.mu.RUnlock()

		statusCode := http.StatusOK
		if overallStatus == StatusUnhealthy {
			statusCode = http.StatusServiceUnavailable
		}

		c.JSON(statusCode, gin.H{
			"ready":     overallStatus == StatusHealthy,
			"status":    overallStatus,
			"checks":    checks,
			"timestamp": time.Now(),
		})
	}
}

func (h *Handler) performHealthChecks(ctx context.Context, verbose bool) Response {
	start := time.Now()
	checks := []Check{}
	overallStatus := StatusHealthy

	// Create a context with timeout for all checks
	checkCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Run all checks concurrently
	var wg sync.WaitGroup
	checkChan := make(chan Check, len(h.checkers))

	h.mu.RLock()
	for name, checker := range h.checkers {
		wg.Add(1)
		go func(n string, c Checker) {
			defer wg.Done()
			check := c.Check(checkCtx)
			check.Name = n
			checkChan <- check
		}(name, checker)
	}
	h.mu.RUnlock()

	// Wait for all checks to complete
	go func() {
		wg.Wait()
		close(checkChan)
	}()

	// Collect results
	for check := range checkChan {
		checks = append(checks, check)

		// Determine overall status
		if check.Status == StatusUnhealthy {
			overallStatus = StatusUnhealthy
		} else if check.Status == StatusDegraded && overallStatus != StatusUnhealthy {
			overallStatus = StatusDegraded
		}
	}

	response := Response{
		Status:    overallStatus,
		Version:   h.version,
		Timestamp: time.Now(),
		Checks:    checks,
		TotalTime: time.Duration(time.Since(start).Milliseconds()),
	}

	// Add system info for verbose requests
	if verbose {
		response.SystemInfo = getSystemInfo()
	}

	return response
}

func (h *Handler) getFromCache(ctx context.Context) (*Response, error) {
	data, err := h.cache.Get(ctx, "health:status").Bytes()
	if err != nil {
		return nil, err
	}

	var response Response
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

func (h *Handler) cacheResponse(ctx context.Context, response Response) {
	data, err := json.Marshal(response)
	if err != nil {
		logger.Error("Failed to marshal health response", zap.Error(err))
		return
	}

	if err := h.cache.Set(ctx, "health:status", data, h.cacheDuration).Err(); err != nil {
		logger.Error("Failed to cache health response", zap.Error(err))
	}
}

// DatabaseChecker checks database health
type DatabaseChecker struct {
	db *gorm.DB
}

func NewDatabaseChecker(db *gorm.DB) *DatabaseChecker {
	return &DatabaseChecker{db: db}
}

func (d *DatabaseChecker) Check(ctx context.Context) Check {
	start := time.Now()

	// Simple query to check connection
	var result int
	err := d.db.WithContext(ctx).Raw("SELECT 1").Scan(&result).Error

	status := StatusHealthy
	message := "Database is healthy"
	details := make(map[string]interface{})

	if err != nil {
		status = StatusUnhealthy
		message = fmt.Sprintf("Database check failed: %v", err)
	} else {
		// Get additional stats
		sqlDB, _ := d.db.DB()
		if sqlDB != nil {
			stats := sqlDB.Stats()
			details["open_connections"] = stats.OpenConnections
			details["in_use"] = stats.InUse
			details["idle"] = stats.Idle
			details["max_open_connections"] = stats.MaxOpenConnections
		}
	}

	return Check{
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Duration:  time.Duration(time.Since(start).Milliseconds()),
		Details:   details,
	}
}

// RedisChecker checks Redis health
type RedisChecker struct {
	client *redis.Client
}

func NewRedisChecker(client *redis.Client) *RedisChecker {
	return &RedisChecker{client: client}
}

func (r *RedisChecker) Check(ctx context.Context) Check {
	start := time.Now()

	// Ping Redis
	_, err := r.client.Ping(ctx).Result()

	status := StatusHealthy
	message := "Redis is healthy"
	details := make(map[string]interface{})

	if err != nil {
		status = StatusUnhealthy
		message = fmt.Sprintf("Redis check failed: %v", err)
	} else {
		// Get Redis info
		info, _ := r.client.Info(ctx).Result()
		if info != "" {
			details["info"] = "available"
		}

		// Get pool stats
		poolStats := r.client.PoolStats()
		details["total_conns"] = poolStats.TotalConns
		details["idle_conns"] = poolStats.IdleConns
		details["stale_conns"] = poolStats.StaleConns
	}

	return Check{
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Duration:  time.Duration(time.Since(start).Milliseconds()),
		Details:   details,
	}
}

// KeycloakChecker checks Keycloak health
type KeycloakChecker struct {
	keycloakClient *auth.KeycloakClient
	baseURL        string
}

func NewKeycloakChecker(client *auth.KeycloakClient, baseURL string) *KeycloakChecker {
	return &KeycloakChecker{
		keycloakClient: client,
		baseURL:        baseURL,
	}
}

func (k *KeycloakChecker) Check(ctx context.Context) Check {
	start := time.Now()

	// Try to get public key (lightweight operation)
	_, err := k.keycloakClient.GetPublicKey(ctx)

	status := StatusHealthy
	message := "Keycloak is healthy"
	details := map[string]interface{}{
		"base_url": k.baseURL,
	}

	if err != nil {
		status = StatusUnhealthy
		message = fmt.Sprintf("Keycloak check failed: %v", err)
	}

	return Check{
		Status:    status,
		Message:   message,
		Timestamp: time.Now(),
		Duration:  time.Duration(time.Since(start).Milliseconds()),
		Details:   details,
	}
}

// getSystemInfo returns system information for verbose health checks
func getSystemInfo() map[string]interface{} {
	// This is a placeholder - implement actual system info gathering
	return map[string]interface{}{
		"go_version": "1.21",
		"os":         "linux",
		"arch":       "amd64",
		"cpus":       4,
		"memory_mb":  8192,
	}
}
