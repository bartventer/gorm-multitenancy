/*
Package mysql provides a [gorm.Dialector] implementation for [MySQL] databases
to support multitenancy in GORM applications, enabling tenant-specific operations
and shared resources management. It includes utilities for registering models,
migrating shared and tenant-specific models, and configuring the database for
tenant-specific operations.

This package follows the "separate databases" approach for multitenancy, which
allows for complete data isolation by utilizing separate databases for each tenant.
This approach ensures maximum security and performance isolation between tenants,
making it suitable for applications with stringent data security requirements.

# URL Format

The URL format for MySQL databases is as follows:

	mysql://user:password@tcp(localhost:3306)/dbname

See the [MySQL connection strings] documentation for more information.

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
[MySQL connection strings]: https://dev.mysql.com/doc/refman/8.4/en/connecting-using-uri-or-key-value-pairs.html#connecting-using-uri
*/
package mysql

import (
	"context"

	"gorm.io/gorm"

	"github.com/bartventer/gorm-multitenancy/mysql/v8/internal/dsn"
	"github.com/bartventer/gorm-multitenancy/mysql/v8/schema"
	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
)

// DriverName is the name of the MySQL driver.
const DriverName = "mysql"

var _ multitenancy.Adapter = new(mysqlAdapter)
var _ driver.DBFactory = new(mysqlAdapter)

// mysqlAdapter is a MySQL-specific implementation of the [driver.DBFactory] interface.
type mysqlAdapter struct{}

func init() { //nolint:gochecknoinits // Required for driver registration.
	multitenancy.Register(DriverName, &mysqlAdapter{})
	multitenancy.Register("mysqlx", &mysqlAdapter{})
}

// AdaptDB implements [multitenancy.Adapter].
func (p *mysqlAdapter) AdaptDB(ctx context.Context, db *gorm.DB) (*multitenancy.DB, error) {
	return multitenancy.NewDB(&mysqlAdapter{}, db), nil
}

// OpenDBURL implements [multitenancy.Adapter].
func (p *mysqlAdapter) OpenDBURL(ctx context.Context, u *driver.URL, opts ...gorm.Option) (*multitenancy.DB, error) {
	urlstr := dsn.StripSchemeFromURL(u.Raw())
	db, err := gorm.Open(Open(urlstr), opts...)
	if err != nil {
		return nil, err
	}
	return p.AdaptDB(ctx, db)
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
