/*
Package postgres provides a [gorm.Dialector] implementation for [PostgreSQL] databases
to support multitenancy in GORM applications, enabling tenant-specific operations
and shared resources management. It includes utilities for registering models,
migrating shared and tenant-specific models, and configuring the database for
tenant-specific operations.

# Model Registration

To register models for multitenancy support, use [RegisterModels]. This should
be done before running any migrations or tenant-specific operations.

# Migration Strategy

To ensure data integrity and schema isolation across tenants, the AutoMigrate method
has been disabled. Instead, use the provided shared and tenant-specific migration methods.
[driver.ErrInvalidMigration] is returned if the migration method is called directly.

# Shared Model Migrations

To migrate shared models, use [MigratePublicSchema].

# Tenant Model Migrations

To migrate tenant-specific models, use [MigrateTenantModels].

# Tenant Offboarding

To clean up the database for a removed tenant, use [DropSchemaForTenant].

# Tenant Context Configuration

To configure the database for operations specific to a tenant, use [schema.SetSearchPath].

# Current Tenant Context

To retrieve the identifier for the current tenant context, use [schema.CurrentSearchPath].

[PostgreSQL]: https://www.postgresql.org
*/
package postgres

import (
	"context"

	"github.com/bartventer/gorm-multitenancy/postgres/v7/schema"
	multitenancy "github.com/bartventer/gorm-multitenancy/v7"
	"github.com/bartventer/gorm-multitenancy/v7/pkg/driver"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// DriverName specifies the PostgreSQL driver name, used for driver registration.
var DriverName = postgres.Dialector{}.Name()

var _ multitenancy.Adapter = new(postgresAdapter)
var _ driver.DBFactory = new(postgresAdapter)

// postgresAdapter is a PostgreSQL-specific implementation of the [driver.DBFactory] interface.
type postgresAdapter struct{}

func init() { //nolint:gochecknoinits // Required for driver registration.
	multitenancy.Register(DriverName, &postgresAdapter{})
}

// AdaptDB implements [multitenancy.Adapter].
func (p *postgresAdapter) AdaptDB(ctx context.Context, db *gorm.DB) (*multitenancy.DB, error) {
	return multitenancy.NewDB(&postgresAdapter{}, db), nil
}

// MigrateSharedModels implements [driver.DBFactory].
func (p *postgresAdapter) MigrateSharedModels(_ context.Context, db *gorm.DB) error {
	return MigratePublicSchema(db)
}

// MigrateTenantModels implements [driver.DBFactory].
func (p *postgresAdapter) MigrateTenantModels(_ context.Context, db *gorm.DB, tenantID string) error {
	return MigrateTenantModels(db, tenantID)
}

// OffboardTenant implements [driver.DBFactory].
func (p *postgresAdapter) OffboardTenant(_ context.Context, db *gorm.DB, tenantID string) error {
	return DropSchemaForTenant(db, tenantID)
}

// RegisterModels implements [driver.DBFactory].
func (p *postgresAdapter) RegisterModels(_ context.Context, db *gorm.DB, models ...driver.TenantTabler) error {
	return RegisterModels(db, models...)
}

// UseTenant implements [driver.DBFactory].
func (p *postgresAdapter) UseTenant(_ context.Context, db *gorm.DB, tenantID string) (func() error, error) {
	return schema.SetSearchPath(db, tenantID)
}

// CurrentTenant implements [driver.DBFactory].
func (p *postgresAdapter) CurrentTenant(_ context.Context, db *gorm.DB) string {
	return schema.CurrentSearchPath(db)
}
