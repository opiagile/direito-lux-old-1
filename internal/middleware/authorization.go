package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/opiagile/direito-lux/internal/authorization"
	"github.com/opiagile/direito-lux/internal/domain"
	"github.com/opiagile/direito-lux/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// AuthorizationMiddleware creates an authorization middleware using OPA
func AuthorizationMiddleware(opaClient *authorization.OPAClient, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip authorization for public endpoints
		if isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Get user context from previous auth middleware
		userID, exists := c.Get("userID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "No user context found"})
			c.Abort()
			return
		}

		tenantName, _ := c.Get("tenant")
		claims, _ := c.Get("claims")
		
		// Build authorization input
		input := buildAuthzInput(c, userID.(string), tenantName.(string), claims)

		// Make authorization decision
		result, err := opaClient.Authorize(c.Request.Context(), input)
		if err != nil {
			logger.Error("Authorization check failed",
				zap.Error(err),
				zap.String("user_id", userID.(string)),
				zap.String("path", c.Request.URL.Path))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization service unavailable"})
			c.Abort()
			return
		}

		// Check authorization result
		if !result.Allow {
			logger.Warn("Authorization denied",
				zap.String("user_id", userID.(string)),
				zap.String("path", c.Request.URL.Path),
				zap.String("reason", result.DenialReason))
			
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied",
				"reason": result.DenialReason,
			})
			c.Abort()
			return
		}

		// Set authorization context for handlers
		c.Set("authz_result", result)
		c.Set("audit_required", result.AuditRequired)
		c.Set("requires_anonymization", result.RequiresAnonymization)

		// Log audit if required
		if result.AuditRequired {
			go logAuditTrail(db, input, result)
		}

		c.Next()
	}
}

// ResourceAuthorizationMiddleware checks authorization for specific resource access
func ResourceAuthorizationMiddleware(opaClient *authorization.OPAClient, resourceType string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID, _ := c.Get("userID")
		tenantName, _ := c.Get("tenant")
		claims, _ := c.Get("claims")
		
		// Get resource ID from path
		resourceID := c.Param("id")
		if resourceID == "" {
			resourceID = c.Param(resourceType + "_id")
		}

		// Build authorization input with resource context
		input := buildAuthzInput(c, userID.(string), tenantName.(string), claims)
		input.Resource.Type = resourceType
		input.Resource.ID = resourceID

		// Get tenant ID for the resource
		if tenantID, exists := c.Get("tenant_id"); exists {
			input.Resource.TenantID = tenantID.(string)
		}

		// Make authorization decision
		result, err := opaClient.Authorize(c.Request.Context(), input)
		if err != nil {
			logger.Error("Resource authorization check failed",
				zap.Error(err),
				zap.String("user_id", userID.(string)),
				zap.String("resource_type", resourceType),
				zap.String("resource_id", resourceID))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authorization service unavailable"})
			c.Abort()
			return
		}

		if !result.Allow {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Access denied to resource",
				"reason": result.DenialReason,
				"resource_type": resourceType,
				"resource_id": resourceID,
			})
			c.Abort()
			return
		}

		// Check for tenant isolation violation
		if result.TenantIsolationViolated {
			logger.Error("Tenant isolation violation detected",
				zap.String("user_id", userID.(string)),
				zap.String("user_tenant", tenantName.(string)),
				zap.String("resource_tenant", input.Resource.TenantID))
			c.JSON(http.StatusForbidden, gin.H{"error": "Tenant isolation violation"})
			c.Abort()
			return
		}

		// Check rate limits
		if result.RateLimitExceeded {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"message": "Your plan's limits have been reached for this resource type",
			})
			c.Abort()
			return
		}

		c.Set("authz_result", result)
		c.Next()
	}
}

// FeatureAuthorizationMiddleware checks if a feature is available for the tenant's plan
func FeatureAuthorizationMiddleware(opaClient *authorization.OPAClient, feature string) gin.HandlerFunc {
	return func(c *gin.Context) {
		claims, _ := c.Get("claims")
		claimsMap := claims.(map[string]interface{})
		
		// Get tenant plan from claims or database
		tenantPlan := "starter" // Default
		if plan, ok := claimsMap["tenant_plan"].(string); ok {
			tenantPlan = plan
		}

		// Check feature availability
		allowed, err := opaClient.CheckFeature(c.Request.Context(), tenantPlan, feature)
		if err != nil {
			logger.Error("Feature check failed",
				zap.Error(err),
				zap.String("feature", feature))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Feature check failed"})
			c.Abort()
			return
		}

		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"error": "Feature not available",
				"message": "This feature is not available in your current plan",
				"feature": feature,
				"plan": tenantPlan,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// buildAuthzInput builds the authorization input from request context
func buildAuthzInput(c *gin.Context, userID, tenantName string, claims interface{}) authorization.AuthzInput {
	claimsMap, _ := claims.(map[string]interface{})
	
	// Extract user info from claims
	email, _ := claimsMap["email"].(string)
	role := extractRole(claimsMap)
	groups := extractGroups(claimsMap)
	
	// Get tenant plan (would typically come from database)
	tenantPlan := "starter"
	if plan, ok := claimsMap["tenant_plan"].(string); ok {
		tenantPlan = plan
	}

	// Build path segments
	pathSegments := strings.Split(strings.Trim(c.Request.URL.Path, "/"), "/")

	// Extract headers (filter sensitive ones)
	headers := make(map[string]string)
	for key, values := range c.Request.Header {
		if !isSensitiveHeader(key) && len(values) > 0 {
			headers[key] = values[0]
		}
	}

	return authorization.AuthzInput{
		User: authorization.User{
			ID:           userID,
			TenantID:     tenantName,
			TenantPlan:   tenantPlan,
			Role:         role,
			Email:        email,
			Groups:       groups,
			IsSuperAdmin: role == "super_admin",
		},
		Resource: authorization.Resource{
			Type:     extractResourceType(c.Request.URL.Path),
			TenantID: tenantName,
		},
		Action:      mapHTTPMethodToAction(c.Request.Method),
		Method:      c.Request.Method,
		Path:        pathSegments,
		PathParams:  extractPathParams(c),
		QueryParams: extractQueryParams(c),
		Headers:     headers,
		ClientIP:    c.ClientIP(),
		Context:     c.GetString("context"),
	}
}

// extractRole extracts user role from claims
func extractRole(claims map[string]interface{}) string {
	// Check realm roles
	if realmAccess, ok := claims["realm_access"].(map[string]interface{}); ok {
		if roles, ok := realmAccess["roles"].([]interface{}); ok {
			for _, role := range roles {
				if roleStr, ok := role.(string); ok {
					// Return first matching role
					switch roleStr {
					case "admin", "lawyer", "secretary", "client", "super_admin":
						return roleStr
					}
				}
			}
		}
	}
	
	// Check resource roles
	if resourceAccess, ok := claims["resource_access"].(map[string]interface{}); ok {
		if clientAccess, ok := resourceAccess["direito-lux-app"].(map[string]interface{}); ok {
			if roles, ok := clientAccess["roles"].([]interface{}); ok {
				for _, role := range roles {
					if roleStr, ok := role.(string); ok {
						return roleStr
					}
				}
			}
		}
	}
	
	return "client" // Default to most restrictive role
}

// extractGroups extracts user groups from claims
func extractGroups(claims map[string]interface{}) []string {
	groups := []string{}
	
	if groupsClaim, ok := claims["groups"].([]interface{}); ok {
		for _, group := range groupsClaim {
			if groupStr, ok := group.(string); ok {
				groups = append(groups, groupStr)
			}
		}
	}
	
	return groups
}

// extractResourceType extracts resource type from path
func extractResourceType(path string) string {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	
	// Common patterns: /api/v1/{resource_type} or /api/v1/{resource_type}/{id}
	if len(segments) >= 3 && segments[0] == "api" {
		return segments[2]
	}
	
	return "unknown"
}

// mapHTTPMethodToAction maps HTTP methods to actions
func mapHTTPMethodToAction(method string) string {
	switch method {
	case "GET":
		return "read"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return strings.ToLower(method)
	}
}

// extractPathParams extracts path parameters
func extractPathParams(c *gin.Context) map[string]string {
	params := make(map[string]string)
	for _, param := range c.Params {
		params[param.Key] = param.Value
	}
	return params
}

// extractQueryParams extracts query parameters
func extractQueryParams(c *gin.Context) map[string]string {
	params := make(map[string]string)
	for key, values := range c.Request.URL.Query() {
		if len(values) > 0 {
			params[key] = values[0]
		}
	}
	return params
}

// isPublicEndpoint checks if endpoint is public
func isPublicEndpoint(path string) bool {
	publicPaths := []string{
		"/health",
		"/api/v1/health",
		"/api/v1/auth/login",
		"/api/v1/auth/refresh",
		"/api/v1/auth/forgot-password",
	}
	
	for _, publicPath := range publicPaths {
		if path == publicPath {
			return true
		}
	}
	
	return false
}

// isSensitiveHeader checks if header contains sensitive information
func isSensitiveHeader(header string) bool {
	sensitive := []string{
		"authorization",
		"cookie",
		"x-api-key",
		"x-auth-token",
	}
	
	headerLower := strings.ToLower(header)
	for _, s := range sensitive {
		if headerLower == s {
			return true
		}
	}
	
	return false
}

// logAuditTrail logs authorization events that require auditing
func logAuditTrail(db *gorm.DB, input authorization.AuthzInput, result *authorization.AuthzResult) {
	audit := &domain.AuditLog{
		TenantID:  uuid.MustParse(input.User.TenantID),
		UserID:    uuid.MustParse(input.User.ID),
		Action:    input.Action,
		Resource:  input.Resource.Type,
		ResourceID: input.Resource.ID,
		IPAddress: input.ClientIP,
		UserAgent: input.Headers["User-Agent"],
		Details: map[string]interface{}{
			"method":        input.Method,
			"path":          input.Path,
			"allowed":       result.Allow,
			"denial_reason": result.DenialReason,
		},
	}
	
	if err := db.Create(audit).Error; err != nil {
		logger.Error("Failed to create audit log",
			zap.Error(err),
			zap.String("user_id", input.User.ID),
			zap.String("action", input.Action))
	}
}