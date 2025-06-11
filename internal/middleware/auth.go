package middleware

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/opiagile/direito-lux/internal/auth"
	"github.com/opiagile/direito-lux/pkg/logger"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

var publicKeyCache *rsa.PublicKey
var publicKeyCacheTime time.Time

// Auth middleware validates JWT tokens
func Auth(keycloakClient *auth.KeycloakClient, redisClient *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractToken(c)
		if token == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No authorization token provided"})
			c.Abort()
			return
		}

		// Check token blacklist in Redis
		ctx := context.Background()
		blacklisted, _ := redisClient.Get(ctx, fmt.Sprintf("blacklist:%s", token)).Result()
		if blacklisted != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Token has been revoked"})
			c.Abort()
			return
		}

		// Check token cache in Redis
		cachedClaims, err := redisClient.Get(ctx, fmt.Sprintf("token:%s", token)).Result()
		if err == nil && cachedClaims != "" {
			// Token is cached and valid
			setClaims(c, cachedClaims)
			c.Next()
			return
		}

		// Validate token with Keycloak
		introspectResult, err := keycloakClient.ValidateToken(ctx, token)
		if err != nil || !*introspectResult.Active {
			logger.Warn("Invalid token",
				zap.String("requestID", c.GetString("requestID")),
				zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Parse JWT to get claims
		claims, err := parseJWT(token, keycloakClient)
		if err != nil {
			logger.Error("Failed to parse JWT",
				zap.String("requestID", c.GetString("requestID")),
				zap.Error(err))
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Failed to parse token"})
			c.Abort()
			return
		}

		// Extract user info from claims
		userID, _ := claims["sub"].(string)
		email, _ := claims["email"].(string)

		// Extract tenant from token
		tenantName, err := auth.ExtractTenantFromToken(claims)
		if err != nil {
			logger.Warn("No tenant found in token",
				zap.String("userID", userID),
				zap.Error(err))
			c.JSON(http.StatusForbidden, gin.H{"error": "No tenant association found"})
			c.Abort()
			return
		}

		// Set user context
		c.Set("userID", userID)
		c.Set("email", email)
		c.Set("tenant", tenantName)
		c.Set("claims", claims)

		// Cache the token (5 minutes)
		redisClient.Set(ctx, fmt.Sprintf("token:%s", token), userID, 5*time.Minute)

		c.Next()
	}
}

// RequireRole checks if user has required role
func RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, exists := c.Get("claims")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{"error": "No claims found"})
			c.Abort()
			return
		}

		claimsMap, ok := claims.(map[string]interface{})
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{"error": "Invalid claims format"})
			c.Abort()
			return
		}

		// Check realm roles
		if realmAccess, ok := claimsMap["realm_access"].(map[string]interface{}); ok {
			if roles, ok := realmAccess["roles"].([]interface{}); ok {
				for _, role := range roles {
					if roleStr, ok := role.(string); ok && roleStr == requiredRole {
						c.Next()
						return
					}
				}
			}
		}

		// Check resource roles
		if resourceAccess, ok := claimsMap["resource_access"].(map[string]interface{}); ok {
			if clientAccess, ok := resourceAccess["direito-lux-app"].(map[string]interface{}); ok {
				if roles, ok := clientAccess["roles"].([]interface{}); ok {
					for _, role := range roles {
						if roleStr, ok := role.(string); ok && roleStr == requiredRole {
							c.Next()
							return
						}
					}
				}
			}
		}

		c.JSON(http.StatusForbidden, gin.H{"error": fmt.Sprintf("Role '%s' required", requiredRole)})
		c.Abort()
	}
}

// extractToken extracts JWT token from Authorization header
func extractToken(c *gin.Context) string {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return ""
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return ""
	}

	return parts[1]
}

// parseJWT parses and validates JWT token
func parseJWT(tokenString string, keycloakClient *auth.KeycloakClient) (map[string]interface{}, error) {
	// Get public key
	publicKey, err := getPublicKey(keycloakClient)
	if err != nil {
		return nil, err
	}

	// Parse token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return publicKey, nil
	})

	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid claims format")
	}

	return claims, nil
}

// getPublicKey gets or caches the realm public key
func getPublicKey(keycloakClient *auth.KeycloakClient) (*rsa.PublicKey, error) {
	// Check cache
	if publicKeyCache != nil && time.Since(publicKeyCacheTime) < 1*time.Hour {
		return publicKeyCache, nil
	}

	// Get public key from Keycloak
	publicKeyStr, err := keycloakClient.GetPublicKey(context.Background())
	if err != nil {
		return nil, err
	}

	// Convert to RSA public key
	publicKey, err := parseRSAPublicKey(publicKeyStr)
	if err != nil {
		return nil, err
	}

	// Cache the key
	publicKeyCache = publicKey
	publicKeyCacheTime = time.Now()

	return publicKey, nil
}

// parseRSAPublicKey converts base64 encoded public key to RSA public key
func parseRSAPublicKey(base64Key string) (*rsa.PublicKey, error) {
	// Decode base64
	decoded, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, err
	}

	// Parse modulus and exponent (simplified - real implementation would need proper ASN.1 parsing)
	// This is a placeholder - you should use proper X.509 parsing
	n := new(big.Int)
	n.SetBytes(decoded)

	publicKey := &rsa.PublicKey{
		N: n,
		E: 65537, // Common RSA exponent
	}

	return publicKey, nil
}

// setClaims sets claims in context (for cached tokens)
func setClaims(c *gin.Context, userID string) {
	c.Set("userID", userID)
	// In a real implementation, you would deserialize full claims from cache
}
