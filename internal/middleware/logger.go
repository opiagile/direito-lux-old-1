package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/opiagile/direito-lux/pkg/logger"
	"go.uber.org/zap"
)

// Logger middleware logs HTTP requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Log request details
		latency := time.Since(start)
		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		errorMessage := c.Errors.ByType(gin.ErrorTypePrivate).String()
		requestID := c.GetString("requestID")

		if query != "" {
			path = path + "?" + query
		}

		fields := []zap.Field{
			zap.String("requestID", requestID),
			zap.String("clientIP", clientIP),
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("statusCode", statusCode),
			zap.Duration("latency", latency),
			zap.String("userAgent", c.Request.UserAgent()),
		}

		if errorMessage != "" {
			fields = append(fields, zap.String("error", errorMessage))
		}

		// Log based on status code
		switch {
		case statusCode >= 500:
			logger.Error("Server error", fields...)
		case statusCode >= 400:
			logger.Warn("Client error", fields...)
		case statusCode >= 300:
			logger.Info("Redirection", fields...)
		default:
			logger.Info("Request completed", fields...)
		}
	}
}