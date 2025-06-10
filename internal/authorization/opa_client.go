package authorization

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/opiagile/direito-lux/pkg/logger"
	"go.uber.org/zap"
)

// OPAClient handles communication with Open Policy Agent
type OPAClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewOPAClient creates a new OPA client
func NewOPAClient(baseURL string) *OPAClient {
	return &OPAClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// AuthzRequest represents an authorization request to OPA
type AuthzRequest struct {
	Input AuthzInput `json:"input"`
}

// AuthzInput contains all the context for authorization decision
type AuthzInput struct {
	User         User                   `json:"user"`
	Resource     Resource               `json:"resource"`
	Action       string                 `json:"action"`
	Method       string                 `json:"method"`
	Path         []string               `json:"path"`
	PathParams   map[string]string      `json:"path_params"`
	QueryParams  map[string]string      `json:"query_params"`
	Headers      map[string]string      `json:"headers"`
	ClientIP     string                 `json:"client_ip"`
	APIKey       string                 `json:"api_key,omitempty"`
	Context      string                 `json:"context,omitempty"`
	Feature      string                 `json:"feature,omitempty"`
}

// User represents user context for OPA
type User struct {
	ID           string   `json:"id"`
	TenantID     string   `json:"tenant_id"`
	TenantPlan   string   `json:"tenant_plan"`
	Role         string   `json:"role"`
	Email        string   `json:"email"`
	Groups       []string `json:"groups,omitempty"`
	IsSuperAdmin bool     `json:"is_super_admin"`
}

// Resource represents the resource being accessed
type Resource struct {
	Type         string                 `json:"type"`
	ID           string                 `json:"id"`
	TenantID     string                 `json:"tenant_id"`
	OwnerID      string                 `json:"owner_id,omitempty"`
	ClientID     string                 `json:"client_id,omitempty"`
	SharedWith   []string               `json:"shared_with,omitempty"`
	Visibility   string                 `json:"visibility,omitempty"`
	Classification string               `json:"classification,omitempty"`
	ContainsPII  bool                   `json:"contains_pii,omitempty"`
	CanBeDeleted bool                   `json:"can_be_deleted,omitempty"`
	AssignedUsers []string              `json:"assigned_users,omitempty"`
	AccessStartTime string              `json:"access_start_time,omitempty"`
	AccessEndTime   string              `json:"access_end_time,omitempty"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// AuthzResponse represents OPA's response
type AuthzResponse struct {
	Result AuthzResult `json:"result"`
}

// AuthzResult contains the authorization decision
type AuthzResult struct {
	Allow              bool     `json:"allow"`
	DenialReason       string   `json:"denial_reason,omitempty"`
	TenantIsolationViolated bool `json:"tenant_isolation_violated,omitempty"`
	RateLimitExceeded  bool     `json:"rate_limit_exceeded,omitempty"`
	AuditRequired      bool     `json:"audit_required,omitempty"`
	RequiresAnonymization bool  `json:"requires_anonymization,omitempty"`
	FeatureAllowed     bool     `json:"feature_allowed,omitempty"`
}

// Authorize makes an authorization decision using OPA
func (c *OPAClient) Authorize(ctx context.Context, input AuthzInput) (*AuthzResult, error) {
	start := time.Now()
	
	// Create request
	req := AuthzRequest{Input: input}
	
	// Marshal request
	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, 
		fmt.Sprintf("%s/v1/data/direitolux/authz", c.baseURL), 
		bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Send request
	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		logger.Error("OPA request failed",
			zap.Error(err),
			zap.String("user_id", input.User.ID),
			zap.String("resource_type", input.Resource.Type))
		return nil, fmt.Errorf("OPA request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("OPA returned status %d", resp.StatusCode)
	}

	// Parse response
	var authzResp AuthzResponse
	if err := json.NewDecoder(resp.Body).Decode(&authzResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Log authorization decision
	logger.Info("Authorization decision",
		zap.String("user_id", input.User.ID),
		zap.String("tenant_id", input.User.TenantID),
		zap.String("resource_type", input.Resource.Type),
		zap.String("action", input.Action),
		zap.Bool("allowed", authzResp.Result.Allow),
		zap.Duration("duration", time.Since(start)))

	return &authzResp.Result, nil
}

// CheckFeature checks if a feature is available for a tenant's plan
func (c *OPAClient) CheckFeature(ctx context.Context, tenantPlan, feature string) (bool, error) {
	input := AuthzInput{
		User: User{
			TenantPlan: tenantPlan,
		},
		Feature: feature,
	}

	result, err := c.Authorize(ctx, input)
	if err != nil {
		return false, err
	}

	return result.FeatureAllowed, nil
}

// LoadPolicies loads or updates OPA policies
func (c *OPAClient) LoadPolicies(ctx context.Context, policies map[string]string) error {
	for name, policy := range policies {
		url := fmt.Sprintf("%s/v1/policies/%s", c.baseURL, name)
		
		req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, 
			bytes.NewReader([]byte(policy)))
		if err != nil {
			return fmt.Errorf("failed to create request for policy %s: %w", name, err)
		}

		req.Header.Set("Content-Type", "text/plain")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("failed to load policy %s: %w", name, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to load policy %s: status %d", name, resp.StatusCode)
		}

		logger.Info("Loaded OPA policy", zap.String("policy", name))
	}

	return nil
}

// LoadData loads data into OPA (plans, tenant settings, etc)
func (c *OPAClient) LoadData(ctx context.Context, path string, data interface{}) error {
	url := fmt.Sprintf("%s/v1/data/%s", c.baseURL, path)
	
	body, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, 
		bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to load data: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("failed to load data: status %d", resp.StatusCode)
	}

	logger.Info("Loaded data into OPA", zap.String("path", path))
	return nil
}

// Health checks OPA health status
func (c *OPAClient) Health(ctx context.Context) error {
	url := fmt.Sprintf("%s/health", c.baseURL)
	
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("OPA health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("OPA unhealthy: status %d", resp.StatusCode)
	}

	return nil
}