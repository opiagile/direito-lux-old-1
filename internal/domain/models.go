package domain

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// BaseModel contains common fields for all models
type BaseModel struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at,omitempty"`
}

// BeforeCreate ensures UUID is generated
func (b *BaseModel) BeforeCreate(tx *gorm.DB) error {
	if b.ID == uuid.Nil {
		b.ID = uuid.New()
	}
	return nil
}

// Tenant represents a law firm or legal professional (multi-tenant isolation)
type Tenant struct {
	BaseModel
	Name            string         `gorm:"not null;uniqueIndex" json:"name"`
	DisplayName     string         `json:"display_name"`
	Domain          string         `gorm:"uniqueIndex" json:"domain,omitempty"`
	KeycloakGroupID string         `gorm:"not null" json:"keycloak_group_id"`
	Plan            Plan           `json:"plan"`
	PlanID          uuid.UUID      `json:"plan_id"`
	Status          TenantStatus   `gorm:"default:'active'" json:"status"`
	Settings        TenantSettings `gorm:"serializer:json" json:"settings"`
	Subscription    *Subscription  `json:"subscription,omitempty"`
}

type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusTrial     TenantStatus = "trial"
	TenantStatusInactive  TenantStatus = "inactive"
)

type TenantSettings struct {
	Language          string                 `json:"language"`
	Timezone          string                 `json:"timezone"`
	DateFormat        string                 `json:"date_format"`
	CurrencyCode      string                 `json:"currency_code"`
	NotificationPrefs NotificationPreferences `json:"notification_prefs"`
	Features          map[string]bool        `json:"features"`
}

type NotificationPreferences struct {
	EmailEnabled    bool     `json:"email_enabled"`
	SMSEnabled      bool     `json:"sms_enabled"`
	WhatsAppEnabled bool     `json:"whatsapp_enabled"`
	TelegramEnabled bool     `json:"telegram_enabled"`
	SlackEnabled    bool     `json:"slack_enabled"`
	EmailTypes      []string `json:"email_types"`
}

// Plan represents subscription plans
type Plan struct {
	BaseModel
	Name         string            `gorm:"not null;uniqueIndex" json:"name"`
	DisplayName  string            `json:"display_name"`
	Description  string            `json:"description"`
	Price        float64           `json:"price"`
	Currency     string            `json:"currency"`
	BillingCycle BillingCycle      `json:"billing_cycle"`
	Features     map[string]interface{} `gorm:"serializer:json" json:"features"`
	Limits       PlanLimits        `gorm:"serializer:json" json:"limits"`
	IsActive     bool              `gorm:"default:true" json:"is_active"`
}

type BillingCycle string

const (
	BillingCycleMonthly  BillingCycle = "monthly"
	BillingCycleYearly   BillingCycle = "yearly"
	BillingCycleOneTime  BillingCycle = "one_time"
)

type PlanLimits struct {
	MaxUsers          int  `json:"max_users"`
	MaxClients        int  `json:"max_clients"`
	MaxCases          int  `json:"max_cases"`
	MaxStorageGB      int  `json:"max_storage_gb"`
	MaxAPICallsMonth  int  `json:"max_api_calls_month"`
	AIRequestsMonth   int  `json:"ai_requests_month"`
	MessagesMonth     int  `json:"messages_month"`
	AllowCustomDomain bool `json:"allow_custom_domain"`
}

// Subscription tracks tenant subscriptions
type Subscription struct {
	BaseModel
	TenantID      uuid.UUID            `gorm:"not null;uniqueIndex" json:"tenant_id"`
	PlanID        uuid.UUID            `gorm:"not null" json:"plan_id"`
	Status        SubscriptionStatus   `json:"status"`
	StartDate     time.Time            `json:"start_date"`
	EndDate       *time.Time           `json:"end_date,omitempty"`
	TrialEndsAt   *time.Time           `json:"trial_ends_at,omitempty"`
	CancelledAt   *time.Time           `json:"cancelled_at,omitempty"`
	PaymentMethod string               `json:"payment_method,omitempty"`
	StripeSubID   string               `json:"stripe_subscription_id,omitempty"`
	Usage         map[string]int       `gorm:"serializer:json" json:"usage"`
}

type SubscriptionStatus string

const (
	SubscriptionStatusActive    SubscriptionStatus = "active"
	SubscriptionStatusTrialing  SubscriptionStatus = "trialing"
	SubscriptionStatusPastDue   SubscriptionStatus = "past_due"
	SubscriptionStatusCanceled  SubscriptionStatus = "canceled"
	SubscriptionStatusUnpaid    SubscriptionStatus = "unpaid"
)

// User represents a user within a tenant
type User struct {
	BaseModel
	KeycloakID    string         `gorm:"not null;uniqueIndex" json:"keycloak_id"`
	TenantID      uuid.UUID      `gorm:"not null;index" json:"tenant_id"`
	Email         string         `gorm:"not null;uniqueIndex" json:"email"`
	FirstName     string         `json:"first_name"`
	LastName      string         `json:"last_name"`
	Role          UserRole       `json:"role"`
	Status        UserStatus     `gorm:"default:'active'" json:"status"`
	Preferences   UserPreferences `gorm:"serializer:json" json:"preferences"`
	LastLoginAt   *time.Time     `json:"last_login_at,omitempty"`
	Tenant        Tenant         `json:"tenant,omitempty"`
}

type UserRole string

const (
	UserRoleAdmin     UserRole = "admin"
	UserRoleLawyer    UserRole = "lawyer"
	UserRoleSecretary UserRole = "secretary"
	UserRoleClient    UserRole = "client"
)

type UserStatus string

const (
	UserStatusActive   UserStatus = "active"
	UserStatusInactive UserStatus = "inactive"
	UserStatusInvited  UserStatus = "invited"
	UserStatusBlocked  UserStatus = "blocked"
)

type UserPreferences struct {
	Language         string `json:"language"`
	Timezone         string `json:"timezone"`
	DateFormat       string `json:"date_format"`
	NotificationsOn  bool   `json:"notifications_on"`
	EmailDigest      string `json:"email_digest"` // daily, weekly, never
}

// AuditLog tracks all actions for compliance
type AuditLog struct {
	ID         uuid.UUID      `gorm:"type:uuid;primary_key" json:"id"`
	TenantID   uuid.UUID      `gorm:"not null;index" json:"tenant_id"`
	UserID     uuid.UUID      `gorm:"index" json:"user_id"`
	Action     string         `gorm:"not null" json:"action"`
	Resource   string         `json:"resource"`
	ResourceID string         `json:"resource_id,omitempty"`
	IPAddress  string         `json:"ip_address"`
	UserAgent  string         `json:"user_agent"`
	Details    map[string]interface{} `gorm:"serializer:json" json:"details,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// BeforeCreate ensures UUID is generated for AuditLog
func (a *AuditLog) BeforeCreate(tx *gorm.DB) error {
	if a.ID == uuid.Nil {
		a.ID = uuid.New()
	}
	return nil
}

// APIKey for service-to-service authentication
type APIKey struct {
	BaseModel
	TenantID    uuid.UUID  `gorm:"not null;index" json:"tenant_id"`
	Name        string     `gorm:"not null" json:"name"`
	Key         string     `gorm:"not null;uniqueIndex" json:"-"`
	Scopes      []string   `gorm:"serializer:json" json:"scopes"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
	LastUsedAt  *time.Time `json:"last_used_at,omitempty"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
}

// === MÓDULO 3 - CONSULTA JURÍDICA ===

// ConsultaProcesso representa uma consulta de processo judicial
type ConsultaProcesso struct {
	ID              string           `json:"id"`
	NumeroProcesso  string           `json:"numero_processo"`
	Tribunal        string           `json:"tribunal"`
	Status          string           `json:"status"`
	DataConsulta    time.Time        `json:"data_consulta"`
	Processo        *ProcessoJudicial `json:"processo,omitempty"`
}

// ProcessoJudicial representa um processo judicial
type ProcessoJudicial struct {
	Numero        string          `json:"numero"`
	Tribunal      string          `json:"tribunal"`
	Classe        string          `json:"classe"`
	Assunto       string          `json:"assunto"`
	Status        string          `json:"status"`
	DataAutuacao  time.Time       `json:"data_autuacao"`
	Partes        []Parte         `json:"partes"`
	Movimentacoes []Movimentacao  `json:"movimentacoes"`
}

// Parte representa uma parte do processo
type Parte struct {
	Nome string `json:"nome"`
	Tipo string `json:"tipo"` // Autor, Réu, etc.
}

// Movimentacao representa uma movimentação processual
type Movimentacao struct {
	Data      time.Time `json:"data"`
	Descricao string    `json:"descricao"`
	Tipo      string    `json:"tipo"`
}

// ConsultaLegislacao representa uma consulta de legislação
type ConsultaLegislacao struct {
	ID           string    `json:"id"`
	Tema         string    `json:"tema"`
	Jurisdicao   string    `json:"jurisdicao"`
	Status       string    `json:"status"`
	DataConsulta time.Time `json:"data_consulta"`
	Leis         []*Lei    `json:"leis,omitempty"`
}

// Lei representa uma lei ou norma jurídica
type Lei struct {
	ID             string    `json:"id"`
	Numero         string    `json:"numero"`
	Nome           string    `json:"nome"`
	Ementa         string    `json:"ementa"`
	DataPublicacao time.Time `json:"data_publicacao"`
	Jurisdicao     string    `json:"jurisdicao"`
	Status         string    `json:"status"`
}

// ConsultaJurisprudencia representa uma consulta de jurisprudência
type ConsultaJurisprudencia struct {
	ID           string     `json:"id"`
	Tema         string     `json:"tema"`
	Tribunal     string     `json:"tribunal"`
	Status       string     `json:"status"`
	DataConsulta time.Time  `json:"data_consulta"`
	Decisoes     []*Decisao `json:"decisoes,omitempty"`
}

// Decisao representa uma decisão judicial
type Decisao struct {
	ID             string    `json:"id"`
	Tribunal       string    `json:"tribunal"`
	Numero         string    `json:"numero"`
	Relator        string    `json:"relator"`
	DataJulgamento time.Time `json:"data_julgamento"`
	Ementa         string    `json:"ementa"`
	Resultado      string    `json:"resultado"`
	Tema           string    `json:"tema"`
}