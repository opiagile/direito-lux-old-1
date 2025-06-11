package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/opiagile/direito-lux/internal/services"
	"github.com/opiagile/direito-lux/pkg/logger"
	"go.uber.org/zap"
)

type TenantHandler struct {
	tenantService *services.TenantService
}

func NewTenantHandler(tenantService *services.TenantService) *TenantHandler {
	return &TenantHandler{
		tenantService: tenantService,
	}
}

// CreateTenant handles POST /api/v1/tenants
func (h *TenantHandler) CreateTenant(c *gin.Context) {
	var req services.CreateTenantRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	tenant, err := h.tenantService.CreateTenant(c.Request.Context(), &req)
	if err != nil {
		logger.Error("Failed to create tenant",
			zap.String("requestID", c.GetString("requestID")),
			zap.Error(err))

		switch err {
		case services.ErrTenantNameExists:
			c.JSON(http.StatusConflict, gin.H{"error": "Tenant name already exists"})
		case services.ErrPlanNotFound:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid plan ID"})
		case services.ErrInvalidTenantData:
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant data"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create tenant"})
		}
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tenant created successfully",
		"data":    tenant,
	})
}

// GetTenant handles GET /api/v1/tenants/:id
func (h *TenantHandler) GetTenant(c *gin.Context) {
	idStr := c.Param("id")
	tenantID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	tenant, err := h.tenantService.GetTenant(c.Request.Context(), tenantID)
	if err != nil {
		if err == services.ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		logger.Error("Failed to get tenant",
			zap.String("tenantID", idStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": tenant})
}

// ListTenants handles GET /api/v1/tenants
func (h *TenantHandler) ListTenants(c *gin.Context) {
	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	offset := (page - 1) * limit

	tenants, total, err := h.tenantService.ListTenants(c.Request.Context(), offset, limit, status)
	if err != nil {
		logger.Error("Failed to list tenants", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to list tenants"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tenants,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// UpdateTenant handles PUT /api/v1/tenants/:id
func (h *TenantHandler) UpdateTenant(c *gin.Context) {
	idStr := c.Param("id")
	tenantID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	tenant, err := h.tenantService.UpdateTenant(c.Request.Context(), tenantID, updates)
	if err != nil {
		if err == services.ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		logger.Error("Failed to update tenant",
			zap.String("tenantID", idStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update tenant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tenant updated successfully",
		"data":    tenant,
	})
}

// GetTenantUsage handles GET /api/v1/tenants/:id/usage
func (h *TenantHandler) GetTenantUsage(c *gin.Context) {
	idStr := c.Param("id")
	tenantID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tenant ID"})
		return
	}

	usage, err := h.tenantService.GetTenantUsage(c.Request.Context(), tenantID)
	if err != nil {
		if err == services.ErrTenantNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Tenant not found"})
			return
		}
		logger.Error("Failed to get tenant usage",
			zap.String("tenantID", idStr),
			zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve usage data"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": usage})
}
