package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/opiagile/direito-lux/internal/auth"
	"github.com/opiagile/direito-lux/internal/domain"
	"github.com/opiagile/direito-lux/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var (
	ErrTenantNotFound    = errors.New("tenant not found")
	ErrTenantNameExists  = errors.New("tenant name already exists")
	ErrPlanNotFound      = errors.New("plan not found")
	ErrInvalidTenantData = errors.New("invalid tenant data")
)

type TenantService struct {
	db             *gorm.DB
	keycloakClient *auth.KeycloakClient
}

func NewTenantService(db *gorm.DB, keycloakClient *auth.KeycloakClient) *TenantService {
	return &TenantService{
		db:             db,
		keycloakClient: keycloakClient,
	}
}

type CreateTenantRequest struct {
	Name        string `json:"name" binding:"required,min=3,max=50"`
	DisplayName string `json:"display_name" binding:"required"`
	Domain      string `json:"domain,omitempty"`
	PlanID      string `json:"plan_id" binding:"required,uuid"`
	AdminUser   struct {
		Email     string `json:"email" binding:"required,email"`
		FirstName string `json:"first_name" binding:"required"`
		LastName  string `json:"last_name" binding:"required"`
		Password  string `json:"password" binding:"required,min=8"`
	} `json:"admin_user" binding:"required"`
	Settings domain.TenantSettings `json:"settings,omitempty"`
}

// CreateTenant creates a new tenant with Keycloak group and admin user
func (ts *TenantService) CreateTenant(ctx context.Context, req *CreateTenantRequest) (*domain.Tenant, error) {
	// Validate tenant name
	req.Name = strings.ToLower(strings.TrimSpace(req.Name))
	if !isValidTenantName(req.Name) {
		return nil, ErrInvalidTenantData
	}

	// Check if tenant name already exists
	var existingTenant domain.Tenant
	if err := ts.db.Where("name = ?", req.Name).First(&existingTenant).Error; err == nil {
		return nil, ErrTenantNameExists
	}

	// Get plan
	var plan domain.Plan
	planID, err := uuid.Parse(req.PlanID)
	if err != nil {
		return nil, ErrInvalidTenantData
	}
	if err := ts.db.Where("id = ? AND is_active = ?", planID, true).First(&plan).Error; err != nil {
		return nil, ErrPlanNotFound
	}

	// Start transaction
	tx := ts.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Create Keycloak group for tenant
	groupID, err := ts.keycloakClient.CreateTenantGroup(ctx, req.Name)
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create Keycloak group: %w", err)
	}

	// Create tenant record
	tenant := &domain.Tenant{
		Name:            req.Name,
		DisplayName:     req.DisplayName,
		Domain:          req.Domain,
		KeycloakGroupID: groupID,
		PlanID:          planID,
		Status:          domain.TenantStatusTrial,
		Settings:        req.Settings,
	}

	// Set default settings
	if tenant.Settings.Language == "" {
		tenant.Settings.Language = "pt-BR"
	}
	if tenant.Settings.Timezone == "" {
		tenant.Settings.Timezone = "America/Sao_Paulo"
	}
	if tenant.Settings.CurrencyCode == "" {
		tenant.Settings.CurrencyCode = "BRL"
	}

	if err := tx.Create(tenant).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create tenant: %w", err)
	}

	// Create subscription
	subscription := &domain.Subscription{
		TenantID:  tenant.ID,
		PlanID:    planID,
		Status:    domain.SubscriptionStatusTrialing,
		StartDate: tenant.CreatedAt,
		Usage:     make(map[string]int),
	}

	if err := tx.Create(subscription).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create subscription: %w", err)
	}

	// Create admin user in Keycloak
	userID, err := ts.keycloakClient.CreateUser(ctx,
		req.AdminUser.Email,
		req.AdminUser.FirstName,
		req.AdminUser.LastName,
		groupID,
		string(domain.UserRoleAdmin))
	if err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create admin user in Keycloak: %w", err)
	}

	// Set password
	if err := ts.keycloakClient.SetUserPassword(ctx, userID, req.AdminUser.Password, false); err != nil {
		tx.Rollback()
		logger.Error("Failed to set admin user password",
			zap.String("userID", userID),
			zap.Error(err))
	}

	// Create user record
	user := &domain.User{
		KeycloakID: userID,
		TenantID:   tenant.ID,
		Email:      req.AdminUser.Email,
		FirstName:  req.AdminUser.FirstName,
		LastName:   req.AdminUser.LastName,
		Role:       domain.UserRoleAdmin,
		Status:     domain.UserStatusActive,
		Preferences: domain.UserPreferences{
			Language:        tenant.Settings.Language,
			Timezone:        tenant.Settings.Timezone,
			NotificationsOn: true,
		},
	}

	if err := tx.Create(user).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create user record: %w", err)
	}

	// Log audit
	audit := &domain.AuditLog{
		TenantID:   tenant.ID,
		UserID:     user.ID,
		Action:     "tenant.created",
		Resource:   "tenant",
		ResourceID: tenant.ID.String(),
		Details: map[string]interface{}{
			"plan_id":    planID.String(),
			"admin_user": req.AdminUser.Email,
		},
	}
	if err := tx.Create(audit).Error; err != nil {
		logger.Error("Failed to create audit log", zap.Error(err))
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.Info("Tenant created successfully",
		zap.String("tenantID", tenant.ID.String()),
		zap.String("name", tenant.Name),
		zap.String("adminUser", req.AdminUser.Email))

	// Load associations
	ts.db.Preload("Plan").Preload("Subscription").First(tenant, tenant.ID)

	return tenant, nil
}

// GetTenant retrieves tenant by ID
func (ts *TenantService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := ts.db.Preload("Plan").Preload("Subscription").Where("id = ?", tenantID).First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

// GetTenantByName retrieves tenant by name
func (ts *TenantService) GetTenantByName(ctx context.Context, name string) (*domain.Tenant, error) {
	var tenant domain.Tenant
	err := ts.db.Preload("Plan").Preload("Subscription").Where("name = ?", name).First(&tenant).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}
	return &tenant, nil
}

// UpdateTenant updates tenant information
func (ts *TenantService) UpdateTenant(ctx context.Context, tenantID uuid.UUID, updates map[string]interface{}) (*domain.Tenant, error) {
	var tenant domain.Tenant
	if err := ts.db.First(&tenant, tenantID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}

	// Only allow certain fields to be updated
	allowedFields := map[string]bool{
		"display_name": true,
		"domain":       true,
		"settings":     true,
		"status":       true,
	}

	filteredUpdates := make(map[string]interface{})
	for key, value := range updates {
		if allowedFields[key] {
			filteredUpdates[key] = value
		}
	}

	if err := ts.db.Model(&tenant).Updates(filteredUpdates).Error; err != nil {
		return nil, err
	}

	// Log audit
	audit := &domain.AuditLog{
		TenantID:   tenantID,
		Action:     "tenant.updated",
		Resource:   "tenant",
		ResourceID: tenantID.String(),
		Details:    filteredUpdates,
	}
	ts.db.Create(audit)

	// Reload with associations
	ts.db.Preload("Plan").Preload("Subscription").First(&tenant, tenantID)

	return &tenant, nil
}

// ListTenants lists all tenants with pagination
func (ts *TenantService) ListTenants(ctx context.Context, offset, limit int, status string) ([]*domain.Tenant, int64, error) {
	var tenants []*domain.Tenant
	var total int64

	query := ts.db.Model(&domain.Tenant{})

	if status != "" {
		query = query.Where("status = ?", status)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.Preload("Plan").
		Preload("Subscription").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&tenants).Error

	return tenants, total, err
}

// GetTenantUsage retrieves current usage statistics for a tenant
func (ts *TenantService) GetTenantUsage(ctx context.Context, tenantID uuid.UUID) (map[string]interface{}, error) {
	var tenant domain.Tenant
	if err := ts.db.Preload("Plan").Preload("Subscription").First(&tenant, tenantID).Error; err != nil {
		return nil, ErrTenantNotFound
	}

	// Count users
	var userCount int64
	ts.db.Model(&domain.User{}).Where("tenant_id = ?", tenantID).Count(&userCount)

	// Get storage usage (placeholder - implement actual storage calculation)
	storageGB := 0.0

	// Get API calls this month (placeholder - implement actual API tracking)
	apiCalls := 0

	usage := map[string]interface{}{
		"users": map[string]interface{}{
			"current": userCount,
			"limit":   tenant.Plan.Limits.MaxUsers,
		},
		"storage_gb": map[string]interface{}{
			"current": storageGB,
			"limit":   tenant.Plan.Limits.MaxStorageGB,
		},
		"api_calls_month": map[string]interface{}{
			"current": apiCalls,
			"limit":   tenant.Plan.Limits.MaxAPICallsMonth,
		},
		"ai_requests_month": map[string]interface{}{
			"current": tenant.Subscription.Usage["ai_requests"],
			"limit":   tenant.Plan.Limits.AIRequestsMonth,
		},
		"messages_month": map[string]interface{}{
			"current": tenant.Subscription.Usage["messages"],
			"limit":   tenant.Plan.Limits.MessagesMonth,
		},
	}

	return usage, nil
}

// isValidTenantName validates tenant name format
func isValidTenantName(name string) bool {
	if len(name) < 3 || len(name) > 50 {
		return false
	}
	// Only allow lowercase letters, numbers, and hyphens
	for _, r := range name {
		if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-') {
			return false
		}
	}
	// Cannot start or end with hyphen
	if name[0] == '-' || name[len(name)-1] == '-' {
		return false
	}
	return true
}
