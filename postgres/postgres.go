/*
Package postgres provides a [gorm.Dialector] implementation for [PostgreSQL] databases
to support multitenancy in GORM applications, enabling tenant-specific operations
and shared resources management. It includes utilities for registering models,
migrating shared and tenant-specific models, and configuring the database for
tenant-specific operations.

This package follows the "shared database, separate schemas" approach for multitenancy,
which allows for data isolation and schema customization per tenant within a single
PostgreSQL database instance. This approach facilitates efficient resource utilization
and simplifies maintenance, making it ideal for applications that require flexible
data isolation without the overhead of managing multiple database instances.

# URL Format

The URL format for PostgreSQL databases is as follows:

	postgres://user:password@localhost:5432/dbname?sslmode=disable

See the [PostgreSQL connection strings] documentation for more information.

# Model Registration

To register models for multitenancy support, use [RegisterModels]. This should
be done before running any migrations or tenant-specific operations.

# Migration Strategy

To ensure data integrity and schema isolation across tenants,[gorm.DB.AutoMigrate] has been
disabled. Instead, use the provided shared and tenant-specific migration methods.
[driver.ErrInvalidMigration] is returned if the `AutoMigrate` method is called directly.

# Concurrent Migrations

To ensure tenant isolation and facilitate concurrent migrations, this package uses
PostgreSQL transaction advisory locks. This mechanism prevents concurrent migrations
from interfering with each other and ensures that only one migration can run at a time.

# Retry Configuration

Exponential backoff retry logic is enabled by default for migrations. To disable retry or
customize the retry behavior, either provide options to [New] or specify options
in the DSN connection string of [Open]. The following options are available:

  - `gmt_disable_retry`: Whether to disable retry. Default is false.
  - `gmt_max_retries`: The maximum number of retry attempts. Default is 6.
  - `gmt_retry_interval`: The initial interval between retry attempts. Default is 2 seconds.
  - `gmt_retry_max_interval`: The maximum interval between retry attempts. Default is 30 seconds.

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
[PostgreSQL connection strings]: https://www.postgresql.org/docs/current/libpq-connect.html#LIBPQ-CONNSTRING-URIS
*/
package postgres

import (
	"context"

	"github.com/bartventer/gorm-multitenancy/postgres/v8/schema"
	multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"gorm.io/gorm"
)

// DriverName is the name of the PostgreSQL driver.
const DriverName = "postgres"

var _ multitenancy.Adapter = new(postgresAdapter)
var _ driver.DBFactory = new(postgresAdapter)

// postgresAdapter is a PostgreSQL-specific implementation of the [driver.DBFactory] interface.
type postgresAdapter struct{}

func init() { //nolint:gochecknoinits // Required for driver registration.
	multitenancy.Register(DriverName, &postgresAdapter{})
	multitenancy.Register("postgresql", &postgresAdapter{}) // Alias for postgres driver.
}

// AdaptDB implements [multitenancy.Adapter].
func (p *postgresAdapter) AdaptDB(ctx context.Context, db *gorm.DB) (*multitenancy.DB, error) {
	return multitenancy.NewDB(&postgresAdapter{}, db), nil
}

// OpenDBURL implements [multitenancy.Adapter].
func (p *postgresAdapter) OpenDBURL(ctx context.Context, u *driver.URL, opts ...gorm.Option) (*multitenancy.DB, error) {
	db, err := gorm.Open(Open(u.Raw()), opts...)
	if err != nil {
		return nil, err
	}
	return p.AdaptDB(ctx, db)
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
