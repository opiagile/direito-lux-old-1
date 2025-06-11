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
	"github.com/opiagile/direito-lux/internal/config"
	"github.com/opiagile/direito-lux/internal/handlers"
	"github.com/opiagile/direito-lux/internal/middleware"
	"github.com/opiagile/direito-lux/internal/services"
	"github.com/opiagile/direito-lux/pkg/circuitbreaker"
	"github.com/opiagile/direito-lux/pkg/logger"
)

// @title Direito Lux - Consulta Jurídica API
// @version 1.0
// @description API para consultas jurídicas com circuit breaker e observabilidade
// @host localhost:9002
// @BasePath /api/v1
func main() {
	// Configuração
	cfg := config.LoadConfig()

	// Logger estruturado
	log := logger.NewLogger(cfg.Logger.Level)

	// Circuit Breaker
	cb := circuitbreaker.NewCircuitBreaker(circuitbreaker.Config{
		Name:        "consulta-juridica",
		MaxRequests: 3,
		Interval:    time.Second * 10,
		Timeout:     time.Second * 60,
		ReadyToTrip: func(counts circuitbreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
	})

	// Services
	consultaService := services.NewConsultaService(log, cb)

	// Handlers
	consultaHandler := handlers.NewConsultaHandler(consultaService, log)

	// Router
	router := gin.New()

	// Middlewares
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger())
	router.Use(middleware.Recovery())
	router.Use(middleware.CORS())

	// Health Check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service":   "Direito Lux - Consulta Jurídica",
			"status":    "healthy",
			"version":   "1.0.0",
			"timestamp": time.Now().Unix(),
			"circuit_breaker": gin.H{
				"name":  cb.Name(),
				"state": cb.State().String(),
			},
		})
	})

	// API Routes
	v1 := router.Group("/api/v1")
	{
		// Consultas jurídicas
		consultas := v1.Group("/consultas")
		{
			consultas.POST("/processos", consultaHandler.ConsultarProcesso)
			consultas.POST("/legislacao", consultaHandler.ConsultarLegislacao)
			consultas.POST("/jurisprudencia", consultaHandler.ConsultarJurisprudencia)
			consultas.GET("/status/:id", consultaHandler.StatusConsulta)
		}

		// Circuit breaker status
		v1.GET("/circuit-breaker/status", func(c *gin.Context) {
			counts := cb.Counts()
			c.JSON(http.StatusOK, gin.H{
				"name":                  cb.Name(),
				"state":                 cb.State().String(),
				"requests":              counts.Requests,
				"total_successes":       counts.TotalSuccesses,
				"total_failures":        counts.TotalFailures,
				"consecutive_successes": counts.ConsecutiveSuccesses,
				"consecutive_failures":  counts.ConsecutiveFailures,
			})
		})
	}

	// Servidor HTTP
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.ConsultaService.Port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		log.Info(fmt.Sprintf("🏛️ Direito Lux - Consulta Jurídica iniciado na porta %s", cfg.ConsultaService.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Falha ao iniciar servidor", "error", err)
		}
	}()

	// Aguardar sinal de interrupção
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info("Desligando servidor...")

	// Shutdown com timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Timeout no shutdown do servidor", "error", err)
	}

	log.Info("Servidor desligado")
}
