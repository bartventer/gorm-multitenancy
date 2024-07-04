// Package driver provides the foundational interfaces for implementing multitenancy support
// within database systems. It outlines the necessary components for managing tenant lifecycles,
// including onboarding, offboarding, and handling shared resources. These interfaces serve as
// a contract for database management systems (DBMS) to ensure consistent multitenant operations,
// abstracting the complexities of tenant-specific data handling. Developers should integrate
// these interfaces with their database solutions to enable scalable and isolated data management
// for each tenant, leveraging the flexibility and power of GORM for ORM operations. This package
// is intended for use by developers implementing multitenant architectures, with the core
// application logic residing elsewhere.
package driver

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

type (
	// DBFactory defines operations for managing the lifecycle of tenants within a multitenant
	// database architecture. It abstracts tenant-specific operations such as onboarding,
	// offboarding, and managing shared resources. Implementations of this interface are designed
	// to integrate multitenancy support with GORM's features, providing a consistent approach
	// across different database backends.
	DBFactory interface {
		// RegisterModels registers GORM model structs for multitenancy support within a specific database.
		// It prepares models for tenant-specific operations and is idempotent. Returns an error if registration fails.
		RegisterModels(ctx context.Context, db *gorm.DB, models ...TenantTabler) error

		// MigrateSharedModels ensures shared data structures are set up and up-to-date within a specific database,
		// maintaining integrity and compatibility of shared data across tenants. Returns an error if migration fails.
		MigrateSharedModels(ctx context.Context, db *gorm.DB) error

		// MigrateTenantModels prepares and updates data structures for a specific tenant within a specific database,
		// handling onboarding and ongoing schema evolution. Returns an error if setup or migration fails.
		MigrateTenantModels(ctx context.Context, db *gorm.DB, tenantID string) error

		// OffboardTenant cleans up the database for a removed tenant within a specific database, supporting clean offboarding.
		// Returns an error if the process fails.
		OffboardTenant(ctx context.Context, db *gorm.DB, tenantID string) error

		// UseTenant configures the database for operations specific to a tenant within a specific database, abstracting
		// database-specific operations for tenant context configuration. Returns a reset function to revert the database context
		// and an error if the operation fails.
		UseTenant(ctx context.Context, db *gorm.DB, tenantID string) (reset func() error, err error)

		// CurrentTenant returns the identifier for the current tenant context within a specific database or an empty string
		// if no context is set.
		CurrentTenant(ctx context.Context, db *gorm.DB) string
	}

	// TenantTabler defines an interface for models within a multi-tenant architecture,
	// extending [schema.Tabler]. Models must define their table name and indicate if they
	// are shared across tenants. Crucial for differentiating between shared and tenant-specific data.
	//
	// Implementations of this interface should return true for [IsSharedModel] if the model is shared
	// across tenants, indicating it does not belong to a single tenant.
	//
	// Example of a shared model:
	//
	// 	type User struct {
	// 		gorm.Model
	// 		Email string
	// 	}
	//
	// 	func (User) TableName() string { return "public.users" }
	// 	func (User) IsSharedModel() bool { return true }
	//
	// Example of a tenant-specific model:
	//
	// 	type Product struct {
	// 		gorm.Model
	// 		TenantID string
	// 		Name     string
	// 	}
	//
	// 	func (Product) TableName() string { return "products" }
	// 	func (Product) IsSharedModel() bool { return false }
	TenantTabler interface {
		schema.Tabler
		// IsSharedModel returns true if the model is shared across tenants, indicating
		// it does not belong to a single tenant.
		IsSharedModel() bool
	}
)
