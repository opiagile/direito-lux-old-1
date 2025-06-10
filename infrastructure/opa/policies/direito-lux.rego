package direitolux.authz

import future.keywords.contains
import future.keywords.if
import future.keywords.in

# Default deny all requests
default allow := false

# Allow health checks without authentication
allow if {
    input.path == ["health"]
    input.method == "GET"
}

allow if {
    input.path == ["api", "v1", "health"]
    input.method == "GET"
}

# Allow authenticated users to access their own profile
allow if {
    input.path == ["api", "v1", "profile"]
    input.method in ["GET", "PUT"]
    input.user.id == input.path_params.user_id
}

# Admin can access everything in their tenant
allow if {
    input.user.role == "admin"
    input.user.tenant_id == input.resource.tenant_id
}

# Lawyer role permissions
allow if {
    input.user.role == "lawyer"
    input.user.tenant_id == input.resource.tenant_id
    lawyer_allowed_resources[input.resource.type]
    lawyer_allowed_actions[input.method]
}

# Secretary role permissions
allow if {
    input.user.role == "secretary"
    input.user.tenant_id == input.resource.tenant_id
    secretary_allowed_resources[input.resource.type]
    secretary_allowed_actions[input.method]
}

# Client role permissions - very restricted
allow if {
    input.user.role == "client"
    input.user.tenant_id == input.resource.tenant_id
    client_allowed_resources[input.resource.type]
    client_allowed_actions[input.method]
    # Clients can only access their own data
    input.resource.client_id == input.user.id
}

# Super admin (platform admin) can access everything
allow if {
    input.user.is_super_admin == true
}

# Define allowed resources per role
lawyer_allowed_resources := {
    "case", "client", "document", "task", "calendar", 
    "report", "billing", "message", "ai_request"
}

secretary_allowed_resources := {
    "case", "client", "document", "task", "calendar", "message"
}

client_allowed_resources := {
    "case", "document", "message", "billing"
}

# Define allowed actions per role
lawyer_allowed_actions := {
    "GET", "POST", "PUT", "DELETE"
}

secretary_allowed_actions := {
    "GET", "POST", "PUT"
}

client_allowed_actions := {
    "GET"
}

# Multi-tenancy enforcement
tenant_isolation_violated if {
    input.user.tenant_id != input.resource.tenant_id
    not input.user.is_super_admin
}

# Rate limiting rules per tenant plan
rate_limit_exceeded if {
    plan := data.plans[input.user.tenant_plan]
    current_usage := data.usage[input.user.tenant_id][input.resource.type]
    current_usage >= plan.limits[input.resource.type]
}

# AI request limits per plan
ai_request_allowed if {
    input.resource.type == "ai_request"
    plan := data.plans[input.user.tenant_plan]
    monthly_usage := data.usage[input.user.tenant_id].ai_requests_month
    monthly_usage < plan.limits.ai_requests_month
}

# Document access rules
document_access_allowed if {
    input.resource.type == "document"
    
    # Owner always has access
    input.resource.owner_id == input.user.id
} else if {
    input.resource.type == "document"
    
    # Shared within tenant based on permissions
    input.user.tenant_id == input.resource.tenant_id
    input.resource.shared_with[input.user.id]
} else if {
    input.resource.type == "document"
    
    # Public documents within tenant
    input.user.tenant_id == input.resource.tenant_id
    input.resource.visibility == "public"
}

# Case access rules
case_access_allowed if {
    input.resource.type == "case"
    
    # Assigned lawyer or secretary
    input.user.id in input.resource.assigned_users
} else if {
    input.resource.type == "case"
    
    # Client of the case
    input.user.role == "client"
    input.user.id == input.resource.client_id
}

# Billing access rules
billing_access_allowed if {
    input.resource.type == "billing"
    input.user.role in ["admin", "lawyer"]
} else if {
    input.resource.type == "billing"
    input.user.role == "client"
    input.resource.client_id == input.user.id
}

# API key validation
api_key_valid if {
    input.api_key
    key := data.api_keys[input.api_key]
    key.tenant_id == input.resource.tenant_id
    key.is_active == true
    key.expires_at > time.now_ns()
    input.path[0] in key.allowed_endpoints
}

# Audit requirements - certain actions must be logged
audit_required if {
    input.resource.type in ["user", "billing", "case", "ai_request"]
} else if {
    input.method in ["DELETE", "PUT"]
} else if {
    input.user.role == "admin"
}

# Data anonymization requirements
requires_anonymization if {
    input.resource.type == "ai_request"
    input.resource.contains_pii == true
} else if {
    input.resource.type == "document"
    input.resource.classification == "confidential"
    input.context == "external_api"
}

# Compliance rules
gdpr_compliant if {
    # User has right to access their own data
    input.action == "read"
    input.resource.owner_id == input.user.id
} else if {
    # User has right to delete their own data
    input.action == "delete" 
    input.resource.owner_id == input.user.id
    input.resource.can_be_deleted == true
} else if {
    # Data export request
    input.action == "export"
    input.resource.owner_id == input.user.id
}

# Time-based access control
time_window_valid if {
    current_time := time.now_ns()
    start_time := time.parse_rfc3339_ns(input.resource.access_start_time)
    end_time := time.parse_rfc3339_ns(input.resource.access_end_time)
    
    current_time >= start_time
    current_time <= end_time
}

# IP-based restrictions
ip_allowed if {
    # Check if IP is in allowed list for tenant
    input.client_ip in data.tenant_settings[input.user.tenant_id].allowed_ips
} else if {
    # No IP restrictions configured
    count(data.tenant_settings[input.user.tenant_id].allowed_ips) == 0
}

# Feature flags per tenant plan
feature_allowed if {
    plan := data.plans[input.user.tenant_plan]
    input.feature in plan.features
}

# Generate detailed denial reasons
denial_reason := reason if {
    not allow
    reasons := [
        ["tenant_isolation", "Access denied: resource belongs to different tenant"] | tenant_isolation_violated,
        ["rate_limit", "Rate limit exceeded for your plan"] | rate_limit_exceeded,
        ["insufficient_role", "Your role does not have permission for this action"] | not role_has_permission,
        ["ip_restricted", "Access denied from this IP address"] | not ip_allowed,
        ["time_window", "Access outside allowed time window"] | not time_window_valid,
        ["feature_disabled", "This feature is not available in your plan"] | not feature_allowed
    ]
    reason := concat(", ", [r[1] | r := reasons[_]; r[0]])
}

# Helper to check role permissions
role_has_permission if {
    input.user.role == "admin"
} else if {
    input.user.role == "lawyer"
    lawyer_allowed_resources[input.resource.type]
    lawyer_allowed_actions[input.method]
} else if {
    input.user.role == "secretary"
    secretary_allowed_resources[input.resource.type]
    secretary_allowed_actions[input.method]
} else if {
    input.user.role == "client"
    client_allowed_resources[input.resource.type]
    client_allowed_actions[input.method]
}