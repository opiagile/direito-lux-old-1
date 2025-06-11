package database

import (
	"fmt"
	"time"

	"github.com/opiagile/direito-lux/internal/domain"
	"github.com/opiagile/direito-lux/pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// MigrationVersion representa uma versão de migration
type MigrationVersion struct {
	ID          uint      `gorm:"primaryKey"`
	Version     string    `gorm:"uniqueIndex;not null"`
	Description string    `gorm:"not null"`
	AppliedAt   time.Time `gorm:"not null"`
	Checksum    string    `gorm:"not null"`
}

// Migration representa uma migration individual
type Migration struct {
	Version     string
	Description string
	Up          func(db *gorm.DB) error
	Down        func(db *gorm.DB) error
	Checksum    string
}

// MigrationManager gerencia as migrations do banco
type MigrationManager struct {
	db         *gorm.DB
	migrations []Migration
}

// NewMigrationManager cria uma nova instância do gerenciador de migrations
func NewMigrationManager(db *gorm.DB) *MigrationManager {
	manager := &MigrationManager{
		db:         db,
		migrations: make([]Migration, 0),
	}

	// Registra todas as migrations
	manager.registerMigrations()

	return manager
}

// registerMigrations registra todas as migrations do sistema
func (m *MigrationManager) registerMigrations() {
	// Migration 001: Criação das tabelas principais
	m.addMigration(Migration{
		Version:     "001_create_initial_tables",
		Description: "Criar tabelas principais: tenants, plans, subscriptions, users, audit_logs, api_keys",
		Checksum:    "sha256:abc123def456", // Em produção, calcular hash real
		Up: func(db *gorm.DB) error {
			return db.AutoMigrate(
				&domain.Tenant{},
				&domain.Plan{},
				&domain.Subscription{},
				&domain.User{},
				&domain.AuditLog{},
				&domain.APIKey{},
			)
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable(
				&domain.APIKey{},
				&domain.AuditLog{},
				&domain.User{},
				&domain.Subscription{},
				&domain.Plan{},
				&domain.Tenant{},
			)
		},
	})

	// Migration 002: Índices de performance
	m.addMigration(Migration{
		Version:     "002_add_performance_indexes",
		Description: "Adicionar índices para melhorar performance das consultas",
		Checksum:    "sha256:def456ghi789",
		Up: func(db *gorm.DB) error {
			// Índices para tabela tenants
			if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status)").Error; err != nil {
				return err
			}
			if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_tenants_created_at ON tenants(created_at)").Error; err != nil {
				return err
			}

			// Índices para tabela users
			if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id)").Error; err != nil {
				return err
			}
			if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_users_email ON users(email)").Error; err != nil {
				return err
			}

			// Índices para tabela audit_logs
			if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_tenant_id ON audit_logs(tenant_id)").Error; err != nil {
				return err
			}
			if err := db.Exec("CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at)").Error; err != nil {
				return err
			}

			return nil
		},
		Down: func(db *gorm.DB) error {
			// Drop índices
			db.Exec("DROP INDEX IF EXISTS idx_tenants_status")
			db.Exec("DROP INDEX IF EXISTS idx_tenants_created_at")
			db.Exec("DROP INDEX IF EXISTS idx_users_tenant_id")
			db.Exec("DROP INDEX IF EXISTS idx_users_email")
			db.Exec("DROP INDEX IF EXISTS idx_audit_logs_tenant_id")
			db.Exec("DROP INDEX IF EXISTS idx_audit_logs_created_at")
			return nil
		},
	})

	// Migration 003: Dados iniciais (seed)
	m.addMigration(Migration{
		Version:     "003_seed_initial_data",
		Description: "Inserir dados iniciais: planos padrão e configurações",
		Checksum:    "sha256:ghi789jkl012",
		Up: func(db *gorm.DB) error {
			return m.seedInitialData(db)
		},
		Down: func(db *gorm.DB) error {
			// Remover dados seed
			db.Exec("DELETE FROM plans WHERE name IN ('starter', 'professional', 'enterprise')")
			return nil
		},
	})
}

// addMigration adiciona uma migration à lista
func (m *MigrationManager) addMigration(migration Migration) {
	m.migrations = append(m.migrations, migration)
}

// RunMigrations executa todas as migrations pendentes
func (m *MigrationManager) RunMigrations() error {
	// Criar tabela de controle de migrations se não existir
	if err := m.db.AutoMigrate(&MigrationVersion{}); err != nil {
		return fmt.Errorf("failed to create migration_versions table: %w", err)
	}

	logger.Info("Starting database migrations")

	for _, migration := range m.migrations {
		applied, err := m.isMigrationApplied(migration.Version)
		if err != nil {
			return fmt.Errorf("failed to check migration status for %s: %w", migration.Version, err)
		}

		if applied {
			logger.Info("Migration already applied", zap.String("version", migration.Version))
			continue
		}

		logger.Info("Applying migration",
			zap.String("version", migration.Version),
			zap.String("description", migration.Description))

		// Executar migration em transação
		err = m.db.Transaction(func(tx *gorm.DB) error {
			if err := migration.Up(tx); err != nil {
				return fmt.Errorf("migration up failed: %w", err)
			}

			// Registrar migration como aplicada
			migrationRecord := MigrationVersion{
				Version:     migration.Version,
				Description: migration.Description,
				AppliedAt:   time.Now(),
				Checksum:    migration.Checksum,
			}

			if err := tx.Create(&migrationRecord).Error; err != nil {
				return fmt.Errorf("failed to record migration: %w", err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("failed to apply migration %s: %w", migration.Version, err)
		}

		logger.Info("Migration applied successfully", zap.String("version", migration.Version))
	}

	logger.Info("Database migrations completed successfully")
	return nil
}

// isMigrationApplied verifica se uma migration já foi aplicada
func (m *MigrationManager) isMigrationApplied(version string) (bool, error) {
	var count int64
	err := m.db.Model(&MigrationVersion{}).Where("version = ?", version).Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// GetAppliedMigrations retorna lista de migrations aplicadas
func (m *MigrationManager) GetAppliedMigrations() ([]MigrationVersion, error) {
	var migrations []MigrationVersion
	err := m.db.Order("applied_at DESC").Find(&migrations).Error
	return migrations, err
}

// RollbackMigration faz rollback de uma migration específica (uso cuidadoso!)
func (m *MigrationManager) RollbackMigration(version string) error {
	// Encontrar a migration
	var migration *Migration
	for _, m := range m.migrations {
		if m.Version == version {
			migration = &m
			break
		}
	}

	if migration == nil {
		return fmt.Errorf("migration %s not found", version)
	}

	logger.Warn("Rolling back migration",
		zap.String("version", version),
		zap.String("description", migration.Description))

	// Executar rollback em transação
	return m.db.Transaction(func(tx *gorm.DB) error {
		if err := migration.Down(tx); err != nil {
			return fmt.Errorf("migration down failed: %w", err)
		}

		// Remover registro da migration
		if err := tx.Where("version = ?", version).Delete(&MigrationVersion{}).Error; err != nil {
			return fmt.Errorf("failed to remove migration record: %w", err)
		}

		return nil
	})
}

// seedInitialData insere dados iniciais no banco
func (m *MigrationManager) seedInitialData(db *gorm.DB) error {
	// Verificar se já existem planos
	var count int64
	db.Model(&domain.Plan{}).Count(&count)
	if count > 0 {
		logger.Info("Initial data already exists, skipping seed")
		return nil
	}

	// Criar planos padrão
	plans := []domain.Plan{
		{
			Name:         "starter",
			DisplayName:  "Starter",
			Description:  "Ideal para advogados autônomos",
			Price:        99.90,
			Currency:     "BRL",
			BillingCycle: domain.BillingCycleMonthly,
			Features: map[string]interface{}{
				"basic_features": true,
				"email_support":  true,
			},
			Limits: domain.PlanLimits{
				MaxUsers:         1,
				MaxClients:       50,
				MaxCases:         100,
				MaxStorageGB:     10,
				MaxAPICallsMonth: 1000,
				AIRequestsMonth:  100,
				MessagesMonth:    500,
			},
		},
		{
			Name:         "professional",
			DisplayName:  "Professional",
			Description:  "Para pequenos escritórios",
			Price:        299.90,
			Currency:     "BRL",
			BillingCycle: domain.BillingCycleMonthly,
			Features: map[string]interface{}{
				"all_features":     true,
				"priority_support": true,
				"api_access":       true,
			},
			Limits: domain.PlanLimits{
				MaxUsers:          5,
				MaxClients:        500,
				MaxCases:          1000,
				MaxStorageGB:      50,
				MaxAPICallsMonth:  10000,
				AIRequestsMonth:   1000,
				MessagesMonth:     5000,
				AllowCustomDomain: true,
			},
		},
		{
			Name:         "enterprise",
			DisplayName:  "Enterprise",
			Description:  "Para grandes escritórios",
			Price:        999.90,
			Currency:     "BRL",
			BillingCycle: domain.BillingCycleMonthly,
			Features: map[string]interface{}{
				"all_features":      true,
				"dedicated_support": true,
				"api_access":        true,
				"white_label":       true,
			},
			Limits: domain.PlanLimits{
				MaxUsers:          -1, // unlimited
				MaxClients:        -1,
				MaxCases:          -1,
				MaxStorageGB:      500,
				MaxAPICallsMonth:  -1,
				AIRequestsMonth:   10000,
				MessagesMonth:     -1,
				AllowCustomDomain: true,
			},
		},
	}

	for _, plan := range plans {
		if err := db.Create(&plan).Error; err != nil {
			return fmt.Errorf("failed to create plan %s: %w", plan.Name, err)
		}
	}

	logger.Info("Initial data seeded successfully", zap.Int("plans_created", len(plans)))
	return nil
}
