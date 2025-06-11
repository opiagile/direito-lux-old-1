package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opiagile/direito-lux/internal/auth"
	"github.com/opiagile/direito-lux/internal/config"
	"github.com/opiagile/direito-lux/internal/domain"
	"github.com/opiagile/direito-lux/internal/handlers"
	"github.com/opiagile/direito-lux/internal/middleware"
	"github.com/opiagile/direito-lux/internal/repository"
	"github.com/opiagile/direito-lux/internal/services"
	"github.com/opiagile/direito-lux/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Errorf("failed to load config: %w", err))
	}

	// Initialize logger
	if err := logger.Init(cfg.Logger.Level, cfg.Logger.Encoding, cfg.Logger.OutputPath); err != nil {
		panic(fmt.Errorf("failed to initialize logger: %w", err))
	}
	defer func() {
		_ = logger.Sync()
	}()

	logger.Info("Starting Direito Lux API",
		zap.String("version", "1.0.0"),
		zap.String("mode", cfg.Server.Mode))

	// Initialize database
	db, err := initDatabase(cfg)
	if err != nil {
		logger.Fatal("Failed to initialize database", zap.Error(err))
	}

	// Initialize Redis
	redisClient := initRedis(cfg)
	defer redisClient.Close()

	// Initialize Keycloak client
	keycloakClient := auth.NewKeycloakClient(&cfg.Keycloak)

	// Initialize repositories
	repos := repository.NewRepositories(db)

	// Initialize services
	tenantService := services.NewTenantService(db, keycloakClient)
	// Add more services as needed

	// Initialize handlers
	tenantHandler := handlers.NewTenantHandler(tenantService)
	// Add more handlers as needed

	// Setup router
	router := setupRouter(cfg, keycloakClient, redisClient, repos, tenantHandler)

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", zap.Error(err))
		}
	}()

	logger.Info("Server started", zap.String("port", cfg.Server.Port))

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server exited")
}

func initDatabase(cfg *config.Config) (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := db.AutoMigrate(
		&domain.Tenant{},
		&domain.Plan{},
		&domain.Subscription{},
		&domain.User{},
		&domain.AuditLog{},
		&domain.APIKey{},
	); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	// Seed initial data
	if err := seedDatabase(db); err != nil {
		logger.Warn("Failed to seed database", zap.Error(err))
	}

	return db, nil
}

func initRedis(cfg *config.Config) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         cfg.GetRedisAddr(),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}

	logger.Info("Connected to Redis")
	return client
}

func setupRouter(
	cfg *config.Config,
	keycloakClient *auth.KeycloakClient,
	redisClient *redis.Client,
	repos *repository.Repositories,
	tenantHandler *handlers.TenantHandler,
) *gin.Engine {
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())
	router.Use(middleware.RequestID())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
			"time":   time.Now().Unix(),
		})
	})

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		public := v1.Group("")
		{
			public.POST("/auth/login", handlers.Login(keycloakClient))
			public.POST("/auth/refresh", handlers.RefreshToken(keycloakClient))
			public.POST("/auth/forgot-password", handlers.ForgotPassword(keycloakClient))
		}

		// Protected routes
		protected := v1.Group("")
		protected.Use(middleware.Auth(keycloakClient, redisClient))
		{
			// Tenant management (admin only)
			admin := protected.Group("")
			admin.Use(middleware.RequireRole("admin"))
			{
				admin.POST("/tenants", tenantHandler.CreateTenant)
				admin.GET("/tenants", tenantHandler.ListTenants)
				admin.GET("/tenants/:id", tenantHandler.GetTenant)
				admin.PUT("/tenants/:id", tenantHandler.UpdateTenant)
				admin.GET("/tenants/:id/usage", tenantHandler.GetTenantUsage)
			}

			// User profile
			protected.GET("/profile", handlers.GetProfile())
			protected.PUT("/profile", handlers.UpdateProfile())

			// Add more protected routes as needed
		}
	}

	return router
}

func seedDatabase(db *gorm.DB) error {
	// Check if plans already exist
	var count int64
	db.Model(&domain.Plan{}).Count(&count)
	if count > 0 {
		return nil
	}

	// Create default plans
	plans := []domain.Plan{
		{
			Name:         "starter",
			DisplayName:  "Starter",
			Description:  "Ideal para advogados autônomos",
			Price:        99.90,
			Currency:     "BRL",
			BillingCycle: domain.BillingCycleMonthly,
			Features: map[string]interface{}{
				"basic_features": true,
				"email_support":  true,
			},
			Limits: domain.PlanLimits{
				MaxUsers:         1,
				MaxClients:       50,
				MaxCases:         100,
				MaxStorageGB:     10,
				MaxAPICallsMonth: 1000,
				AIRequestsMonth:  100,
				MessagesMonth:    500,
			},
		},
		{
			Name:         "professional",
			DisplayName:  "Professional",
			Description:  "Para pequenos escritórios",
			Price:        299.90,
			Currency:     "BRL",
			BillingCycle: domain.BillingCycleMonthly,
			Features: map[string]interface{}{
				"all_features":     true,
				"priority_support": true,
				"api_access":       true,
			},
			Limits: domain.PlanLimits{
				MaxUsers:          5,
				MaxClients:        500,
				MaxCases:          1000,
				MaxStorageGB:      50,
				MaxAPICallsMonth:  10000,
				AIRequestsMonth:   1000,
				MessagesMonth:     5000,
				AllowCustomDomain: true,
			},
		},
		{
			Name:         "enterprise",
			DisplayName:  "Enterprise",
			Description:  "Para grandes escritórios",
			Price:        999.90,
			Currency:     "BRL",
			BillingCycle: domain.BillingCycleMonthly,
			Features: map[string]interface{}{
				"all_features":      true,
				"dedicated_support": true,
				"api_access":        true,
				"white_label":       true,
			},
			Limits: domain.PlanLimits{
				MaxUsers:          -1, // unlimited
				MaxClients:        -1,
				MaxCases:          -1,
				MaxStorageGB:      500,
				MaxAPICallsMonth:  -1,
				AIRequestsMonth:   10000,
				MessagesMonth:     -1,
				AllowCustomDomain: true,
			},
		},
	}

	for _, plan := range plans {
		if err := db.Create(&plan).Error; err != nil {
			return err
		}
	}

	logger.Info("Database seeded with default plans")
	return nil
}
