package repository

import "gorm.io/gorm"

// Repositories holds all repository instances
type Repositories struct {
	db *gorm.DB
	// Add specific repositories here as needed
	// Tenant TenantRepository
	// User   UserRepository
}

// NewRepositories creates a new instance of repositories
func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		db: db,
	}
}