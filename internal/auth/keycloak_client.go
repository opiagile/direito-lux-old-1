package auth

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/Nerzal/gocloak/v13"
	"github.com/opiagile/direito-lux/internal/config"
	"github.com/opiagile/direito-lux/pkg/logger"
	"go.uber.org/zap"
)

type KeycloakClient struct {
	client        *gocloak.GoCloak
	config        *config.KeycloakConfig
	adminToken    *gocloak.JWT
	tokenMutex    sync.RWMutex
	lastTokenTime time.Time
}

func NewKeycloakClient(cfg *config.KeycloakConfig) *KeycloakClient {
	client := gocloak.NewClient(cfg.BaseURL)

	return &KeycloakClient{
		client: client,
		config: cfg,
	}
}

// getAdminToken gets or refreshes admin token
func (kc *KeycloakClient) getAdminToken(ctx context.Context) (*gocloak.JWT, error) {
	kc.tokenMutex.RLock()
	if kc.adminToken != nil && time.Since(kc.lastTokenTime) < time.Duration(kc.adminToken.ExpiresIn-60)*time.Second {
		token := kc.adminToken
		kc.tokenMutex.RUnlock()
		return token, nil
	}
	kc.tokenMutex.RUnlock()

	kc.tokenMutex.Lock()
	defer kc.tokenMutex.Unlock()

	// Double-check after acquiring write lock
	if kc.adminToken != nil && time.Since(kc.lastTokenTime) < time.Duration(kc.adminToken.ExpiresIn-60)*time.Second {
		return kc.adminToken, nil
	}

	token, err := kc.client.LoginAdmin(ctx, kc.config.AdminUser, kc.config.AdminPass, "master")
	if err != nil {
		return nil, fmt.Errorf("failed to login admin: %w", err)
	}

	kc.adminToken = token
	kc.lastTokenTime = time.Now()

	return token, nil
}

// CreateTenantGroup creates a new group for tenant isolation
func (kc *KeycloakClient) CreateTenantGroup(ctx context.Context, tenantName string) (string, error) {
	token, err := kc.getAdminToken(ctx)
	if err != nil {
		return "", err
	}

	group := gocloak.Group{
		Name: gocloak.StringP(tenantName),
		Path: gocloak.StringP("/" + tenantName),
		Attributes: &map[string][]string{
			"tenant": {tenantName},
			"type":   {"tenant"},
		},
	}

	groupID, err := kc.client.CreateGroup(ctx, token.AccessToken, kc.config.Realm, group)
	if err != nil {
		return "", fmt.Errorf("failed to create tenant group: %w", err)
	}

	logger.Info("Created tenant group",
		zap.String("tenant", tenantName),
		zap.String("groupID", groupID))

	return groupID, nil
}

// CreateUser creates a new user in Keycloak and assigns to tenant group
func (kc *KeycloakClient) CreateUser(ctx context.Context, email, firstName, lastName, tenantGroupID string, role string) (string, error) {
	token, err := kc.getAdminToken(ctx)
	if err != nil {
		return "", err
	}

	user := gocloak.User{
		Username:      gocloak.StringP(email),
		Email:         gocloak.StringP(email),
		FirstName:     gocloak.StringP(firstName),
		LastName:      gocloak.StringP(lastName),
		Enabled:       gocloak.BoolP(true),
		EmailVerified: gocloak.BoolP(false),
		Groups:        &[]string{tenantGroupID},
		Attributes: &map[string][]string{
			"tenant_group_id": {tenantGroupID},
			"role":            {role},
		},
	}

	userID, err := kc.client.CreateUser(ctx, token.AccessToken, kc.config.Realm, user)
	if err != nil {
		return "", fmt.Errorf("failed to create user: %w", err)
	}

	// Send verification email
	err = kc.client.SendVerifyEmail(ctx, token.AccessToken, userID, kc.config.Realm, gocloak.SendVerificationMailParams{})
	if err != nil {
		logger.Warn("Failed to send verification email",
			zap.String("userID", userID),
			zap.Error(err))
	}

	// Assign role
	if err := kc.assignRole(ctx, token.AccessToken, userID, role); err != nil {
		logger.Error("Failed to assign role",
			zap.String("userID", userID),
			zap.String("role", role),
			zap.Error(err))
	}

	logger.Info("Created user",
		zap.String("email", email),
		zap.String("userID", userID),
		zap.String("role", role))

	return userID, nil
}

// assignRole assigns a realm role to a user
func (kc *KeycloakClient) assignRole(ctx context.Context, token, userID, roleName string) error {
	// Get role
	role, err := kc.client.GetRealmRole(ctx, token, kc.config.Realm, roleName)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Assign role to user
	err = kc.client.AddRealmRoleToUser(ctx, token, kc.config.Realm, userID, []gocloak.Role{*role})
	if err != nil {
		return fmt.Errorf("failed to assign role: %w", err)
	}

	return nil
}

// GetUser retrieves user by ID
func (kc *KeycloakClient) GetUser(ctx context.Context, userID string) (*gocloak.User, error) {
	token, err := kc.getAdminToken(ctx)
	if err != nil {
		return nil, err
	}

	user, err := kc.client.GetUserByID(ctx, token.AccessToken, kc.config.Realm, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// UpdateUser updates user information
func (kc *KeycloakClient) UpdateUser(ctx context.Context, userID string, user gocloak.User) error {
	token, err := kc.getAdminToken(ctx)
	if err != nil {
		return err
	}

	err = kc.client.UpdateUser(ctx, token.AccessToken, kc.config.Realm, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// DeleteUser deletes a user from Keycloak
func (kc *KeycloakClient) DeleteUser(ctx context.Context, userID string) error {
	token, err := kc.getAdminToken(ctx)
	if err != nil {
		return err
	}

	err = kc.client.DeleteUser(ctx, token.AccessToken, kc.config.Realm, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}

// GetUsersByTenant gets all users belonging to a tenant group
func (kc *KeycloakClient) GetUsersByTenant(ctx context.Context, tenantGroupID string) ([]*gocloak.User, error) {
	token, err := kc.getAdminToken(ctx)
	if err != nil {
		return nil, err
	}

	// Get group members
	users, err := kc.client.GetGroupMembers(ctx, token.AccessToken, kc.config.Realm, tenantGroupID, gocloak.GetGroupsParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}

	return users, nil
}

// ValidateToken validates a JWT token
func (kc *KeycloakClient) ValidateToken(ctx context.Context, tokenString string) (*gocloak.IntroSpectTokenResult, error) {
	result, err := kc.client.RetrospectToken(ctx, tokenString, kc.config.ClientID, kc.config.ClientSecret, kc.config.Realm)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %w", err)
	}

	return result, nil
}

// GetPublicKey gets the realm public key for JWT verification
func (kc *KeycloakClient) GetPublicKey(ctx context.Context) (string, error) {
	// Get realm certs endpoint
	certs, err := kc.client.GetCerts(ctx, kc.config.Realm)
	if err != nil {
		return "", fmt.Errorf("failed to get realm certificates: %w", err)
	}

	if certs.Keys == nil || len(*certs.Keys) == 0 {
		return "", fmt.Errorf("no public keys found for realm")
	}

	// Get the first key (usually the active one)
	firstKey := (*certs.Keys)[0]
	if firstKey.X5c == nil || len(*firstKey.X5c) == 0 {
		return "", fmt.Errorf("no x5c certificate found in key")
	}

	return (*firstKey.X5c)[0], nil
}

// ResetPassword sends password reset email
func (kc *KeycloakClient) ResetPassword(ctx context.Context, userID string) error {
	token, err := kc.getAdminToken(ctx)
	if err != nil {
		return err
	}

	err = kc.client.SendVerifyEmail(ctx, token.AccessToken, userID, kc.config.Realm, gocloak.SendVerificationMailParams{})
	if err != nil {
		return fmt.Errorf("failed to send reset password email: %w", err)
	}

	return nil
}

// SetUserPassword sets user password directly (for admin operations)
func (kc *KeycloakClient) SetUserPassword(ctx context.Context, userID, password string, temporary bool) error {
	token, err := kc.getAdminToken(ctx)
	if err != nil {
		return err
	}

	err = kc.client.SetPassword(ctx, token.AccessToken, userID, kc.config.Realm, password, temporary)
	if err != nil {
		return fmt.Errorf("failed to set password: %w", err)
	}

	return nil
}

// GetUserRoles gets all roles assigned to a user
func (kc *KeycloakClient) GetUserRoles(ctx context.Context, userID string) ([]*gocloak.Role, error) {
	token, err := kc.getAdminToken(ctx)
	if err != nil {
		return nil, err
	}

	roles, err := kc.client.GetRealmRolesByUserID(ctx, token.AccessToken, kc.config.Realm, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", err)
	}

	return roles, nil
}

// ExtractTenantFromToken extracts tenant ID from JWT claims
func ExtractTenantFromToken(claims map[string]interface{}) (string, error) {
	// Check for tenant in resource access
	if resourceAccess, ok := claims["resource_access"].(map[string]interface{}); ok {
		if clientAccess, ok := resourceAccess["direito-lux-app"].(map[string]interface{}); ok {
			if roles, ok := clientAccess["roles"].([]interface{}); ok {
				for _, role := range roles {
					if roleStr, ok := role.(string); ok && strings.HasPrefix(roleStr, "tenant:") {
						return strings.TrimPrefix(roleStr, "tenant:"), nil
					}
				}
			}
		}
	}

	// Check for tenant in groups
	if groups, ok := claims["groups"].([]interface{}); ok {
		for _, group := range groups {
			if groupStr, ok := group.(string); ok && strings.HasPrefix(groupStr, "/") {
				parts := strings.Split(groupStr, "/")
				if len(parts) > 1 {
					return parts[1], nil
				}
			}
		}
	}

	return "", fmt.Errorf("no tenant found in token")
}
