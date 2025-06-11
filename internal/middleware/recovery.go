package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/opiagile/direito-lux/pkg/logger"
	"go.uber.org/zap"
)

// Recovery middleware recovers from panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				logger.Error("Panic recovered",
					zap.Any("error", err),
					zap.String("path", c.Request.URL.Path),
					zap.String("method", c.Request.Method),
					zap.String("clientIP", c.ClientIP()),
					zap.String("requestID", c.GetString("requestID")),
				)

				c.JSON(http.StatusInternalServerError, gin.H{
					"error":     "Internal server error",
					"requestID": c.GetString("requestID"),
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
