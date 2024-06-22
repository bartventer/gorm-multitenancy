package postgres

import "gorm.io/gorm"

// TenantTabler is the interface for tenant tables.
type TenantTabler interface {
	// IsTenantTable returns true if the table is a tenant table
	IsTenantTable() bool
}

// MultitenancyMigrator is the interface for the postgres migrator with multitenancy support.
type MultitenancyMigrator interface {
	gorm.Migrator

	// CreateSchemaForTenant creates the schema for the tenant, and migrates the private tables
	//
	// Parameters:
	// 	- tenant: the tenant's schema name
	//
	// Returns:
	// 	- error: the error if any
	CreateSchemaForTenant(tenant string) error
	// DropSchemaForTenant drops the schema for the tenant (CASCADING tables)
	//
	// Parameters:
	// 	- tenant: the tenant's schema name
	//
	// Returns:
	// 	- error: the error if any
	DropSchemaForTenant(tenant string) error
	// MigratePublicSchema migrates the public tables
	MigratePublicSchema() error
}
