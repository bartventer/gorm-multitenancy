/*
Package mysql provides a [gorm.Dialector] implementation for [MySQL] databases
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

To migrate shared models, use [MigrateSharedModels].

# Tenant Model Migrations

To migrate tenant-specific models, use [MigrateTenantModels].

# Tenant Offboarding

To clean up the database for a removed tenant, use [DropDatabaseForTenant].

# Tenant Context Configuration

To configure the database for operations specific to a tenant, use [schema.UseDatabase].

[MySQL]: https://www.mysql.com
*/
package mysql

import (
	"context"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"github.com/bartventer/gorm-multitenancy/mysql/v8/schema"
	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
)

// DriverName specifies the MySQL driver name, used for driver registration.
var DriverName = mysql.Dialector{}.Name()

var _ multitenancy.Adapter = new(mysqlAdapter)
var _ driver.DBFactory = new(mysqlAdapter)

// mysqlAdapter is a MySQL-specific implementation of the [driver.DBFactory] interface.
type mysqlAdapter struct{}

func init() { //nolint:gochecknoinits // Required for driver registration.
	multitenancy.Register(DriverName, &mysqlAdapter{})
}

// AdaptDB implements [multitenancy.Adapter].
func (p *mysqlAdapter) AdaptDB(ctx context.Context, db *gorm.DB) (*multitenancy.DB, error) {
	return multitenancy.NewDB(&mysqlAdapter{}, db), nil
}

// MigrateSharedModels implements [driver.DBFactory].
func (p *mysqlAdapter) MigrateSharedModels(_ context.Context, db *gorm.DB) error {
	return MigrateSharedModels(db)
}

// MigrateTenantModels implements [driver.DBFactory].
func (p *mysqlAdapter) MigrateTenantModels(_ context.Context, db *gorm.DB, tenantID string) error {
	return MigrateTenantModels(db, tenantID)
}

// OffboardTenant implements [driver.DBFactory].
func (p *mysqlAdapter) OffboardTenant(_ context.Context, db *gorm.DB, tenantID string) error {
	return DropDatabaseForTenant(db, tenantID)
}

// RegisterModels implements [driver.DBFactory].
func (p *mysqlAdapter) RegisterModels(_ context.Context, db *gorm.DB, models ...driver.TenantTabler) error {
	return RegisterModels(db, models...)
}

// UseTenant implements [driver.DBFactory].
func (p *mysqlAdapter) UseTenant(ctx context.Context, db *gorm.DB, tenantID string) (func() error, error) {
	return schema.UseDatabase(db, tenantID)
}

// CurrentTenant implements [driver.DBFactory].
func (p *mysqlAdapter) CurrentTenant(ctx context.Context, db *gorm.DB) string {
	return db.Migrator().CurrentDatabase()
}
