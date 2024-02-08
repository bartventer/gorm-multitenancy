package postgres

import (
	multitenancy "github.com/bartventer/gorm-multitenancy/v3"
)

// MultitenancyMigrator is the interface for the postgres migrator with multitenancy support
type MultitenancyMigrator interface {
	multitenancy.Migrator

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
