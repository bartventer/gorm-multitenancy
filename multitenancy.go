/*
Package multitenancy provides a framework for implementing multitenancy in Go applications. It
simplifies the development and management of multi-tenant applications by offering functionalities
for tenant-specific and shared model migrations, alongside thorough tenant management. Central to
this package is its ability to abstract multitenancy complexities, presenting a unified,
database-agnostic API that integrates seamlessly with GORM.

The framework supports two primary multitenancy strategies:
  - Shared Database, Separate Schemas: This approach allows for data isolation and schema
    customization per tenant within a single database instance, simplifying maintenance and
    resource utilization.
  - Separate Databases: This strategy ensures complete data isolation by utilizing separate
    databases for each tenant, making it suitable for applications with stringent data security
    requirements.

# Usage

The following sections provide an overview of the package's features and usage instructions for
implementing multitenancy in Go applications.

# Opening a Database Connection

The package supports multitenancy for both PostgreSQL and MySQL databases, offering three methods
for establishing a new database connection with multitenancy support:

# Approach 1: OpenDB with URL (Recommended for Most Users)

[OpenDB] allows opening a database connection using a URL-like DSN string, providing a flexible
and easy way to switch between drivers. This method abstracts the underlying driver mechanics,
offering a straightforward connection process and a unified, database-agnostic API through the
returned [*DB] instance, which embeds the [gorm.DB] instance.

For constructing the DSN string, refer to the driver-specific documentation for the required
parameters and formats.

	import (
	    _ "github.com/bartventer/gorm-multitenancy/<driver>/v8"
	    multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	)

	func main() {
	    url := "<driver>://user:password@host:port/dbname"
	    db, err := multitenancy.OpenDB(context.Background(), url)
	    if err != nil {...}
		db.RegisterModels(ctx, ...) // Access to a database-agnostic API with GORM features
	}

This approach is useful for applications that need to dynamically switch between
different database drivers or configurations without changing the codebase.

Postgres:

	import (
	    _ "github.com/bartventer/gorm-multitenancy/postgres/v8"
	    multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	)

	func main() {
	    url := "postgres://user:password@localhost:5432/dbname?sslmode=disable"
	    db, err := multitenancy.OpenDB(context.Background(), url)
	    if err != nil {...}
	}

MySQL:

	import (
	    _ "github.com/bartventer/gorm-multitenancy/mysql/v8"
	    multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	)

	func main() {
	    url := "mysql://user:password@tcp(localhost:3306)/dbname"
	    db, err := multitenancy.OpenDB(context.Background(), url)
	    if err != nil {...}
	}

# Approach 2: Unified API

[Open] with a supported driver offers a unified, database-agnostic API for managing tenant-specific
and shared data, embedding the [gorm.DB] instance. This method allows developers to initialize and
configure the dialect themselves before opening the database connection, providing granular control
over the database connection and configuration. It facilitates seamless switching between database
drivers while maintaining access to GORM's full functionality.

	import (
	    multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	    "github.com/bartventer/gorm-multitenancy/<driver>/v8"
	)

	func main() {
	    dsn := "<driver-specific DSN>"
	    db, err := multitenancy.Open(<driver>.Open(dsn))
	    if err != nil {...}
	    db.RegisterModels(ctx, ...) // Access to a database-agnostic API with GORM features
	}

Postgres:

	import (
	    multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	    "github.com/bartventer/gorm-multitenancy/postgres/v8"
	)

	func main() {
	    dsn := "postgres://user:password@localhost:5432/dbname?sslmode=disable"
	    db, err := multitenancy.Open(postgres.Open(dsn))
	}

MySQL:

	    import (
	        multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	        "github.com/bartventer/gorm-multitenancy/mysql/v8"
	    )

	    func main() {
	        dsn := "user:password@tcp(localhost:3306)/dbname"
			db, err := multitenancy.Open(mysql.Open(dsn))
	    }

# Approach 3: Direct Driver API

For direct access to the [gorm.DB] API and multitenancy features for specific tasks, this approach
allows invoking driver-specific functions directly. It provides a lower level of abstraction,
requiring manual management of database-specific operations. Prior to version 8, this was the
only method available for using the package.

Postgres:

	import (
	    "github.com/bartventer/gorm-multitenancy/postgres/v8"
	    "gorm.io/gorm"
	)

	func main() {
	    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	    // Directly call driver-specific functions
	    postgres.RegisterModels(db, ...)
	}

MySQL:

	import (
	    "github.com/bartventer/gorm-multitenancy/mysql/v8"
	    "gorm.io/gorm"
	)

	func main() {
	    db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	    // Directly call driver-specific functions
	    mysql.RegisterModels(db, ...)
	}

# Declaring Models

All models must implement [driver.TenantTabler], which extends GORM's Tabler interface. This
extension allows models to define their table name and indicate whether they are shared across
tenants.

Public Models:

These are models which are shared across tenants.
[driver.TenantTabler.TableName] should return the table name prefixed with 'public.'.
[driver.TenantTabler.IsSharedModel] should return true.

	type Tenant struct { multitenancy.TenantModel}
	func (Tenant) TableName() string   { return "public.tenants" } // note the 'public.' prefix
	func (Tenant) IsSharedModel() bool { return true }

Tenant-Specific Models:

These models are specific to a single tenant and should not be shared across tenants.
[driver.TenantTabler.TableName] should return the table name without any prefix.
[driver.TenantTabler.IsSharedModel] should return false.

	type Book struct {
		gorm.Model
		Title        string
		TenantSchema string
		Tenant       Tenant `gorm:"foreignKey:TenantSchema;references:SchemaName"`
	}

	func (Book) TableName() string   { return "books" } // no 'public.' prefix
	func (Book) IsSharedModel() bool { return false }

Tenant Model:

This package provides a [TenantModel] struct that can be embedded in any public model that requires
tenant scoping, enriching it with essential fields for managing tenant-specific information. This
structure incorporates fields for the tenant's domain URL and schema name, facilitating the linkage
of tenant-specific models to their respective schemas.

	type Tenant struct { multitenancy.TenantModel }

	func (Tenant) TableName() string   { return "public.tenants" }
	func (Tenant) IsSharedModel() bool { return true }

# Model Registration

Before performing any migrations or operations on tenant-specific models, the models
must be registered with the DB instance using [DB.RegisterModels].

	import (
		multitenancy "github.com/bartventer/gorm-multitenancy/v8"
		"github.com/bartventer/gorm-multitenancy/postgres/v8"
	)

	func main() {
		db, err := multitenancy.Open(postgres.Open(dsn))
		if err != nil {...}
		db.RegisterModels(ctx, &Tenant{}, &Book{})
	}

Postgres Adapter:

Use [postgres.RegisterModels] to register models.

	import "github.com/bartventer/gorm-multitenancy/postgres/v8"

	postgres.RegisterModels(db, &Tenant{}, &Book{})

MySQL Adapter:

Use [mysql.RegisterModels] to register models.

	import "github.com/bartventer/gorm-multitenancy/mysql/v8"

	mysql.RegisterModels(db, &Tenant{}, &Book{})

# Migration Strategy

To ensure data integrity and schema isolation across tenants, the AutoMigrate method
has been disabled. Instead, use the provided shared and tenant-specific migration methods.
[driver.ErrInvalidMigration] is returned if the AutoMigrate method is called directly.

# Shared Model Migrations

After registering models, shared models are migrated using [DB.MigrateSharedModels].

	import (
		"context"
		multitenancy "github.com/bartventer/gorm-multitenancy/v8"
		"github.com/bartventer/gorm-multitenancy/postgres/v8"
	)

	func main() {
		db, err := multitenancy.Open(postgres.Open(dsn))
		if err != nil {...}
		db.RegisterModels(ctx, &Tenant{}, &Book{})
		db.MigrateSharedModels(ctx)
	}

Postgres Adapter:

Use [postgres.MigrateSharedModels] to migrate shared models.

	import "github.com/bartventer/gorm-multitenancy/postgres/v8"

	postgres.MigrateSharedModels(db)

MySQL Adapter:

Use [mysql.MigrateSharedModels] to migrate shared models.

	import "github.com/bartventer/gorm-multitenancy/mysql/v8"

	mysql.MigrateSharedModels(db)

# Tenant-Specific Model Migrations

After registering models, tenant-specific models are migrated using [DB.MigrateTenantModels].

	import (
		"context"
		multitenancy "github.com/bartventer/gorm-multitenancy/v8"
		"github.com/bartventer/gorm-multitenancy/postgres/v8"
	)

	func main() {
		db, err := multitenancy.Open(postgres.Open(dsn))
		if err != nil {...}
		db.RegisterModels(ctx, &Tenant{}, &Book{})
		db.MigrateSharedModels(ctx)
		// Assuming we have a tenant with schema name 'tenant1'
		db.MigrateTenantModels(ctx, "tenant1")
	}

Postgres Adapter:

Use [postgres.MigrateTenantModels] to migrate tenant-specific models.

	import "github.com/bartventer/gorm-multitenancy/postgres/v8"

	postgres.MigrateTenantModels(db, "tenant1")

MySQL Adapter:

Use [mysql.MigrateTenantModels] to migrate tenant-specific models.

	import "github.com/bartventer/gorm-multitenancy/mysql/v8"

	mysql.MigrateTenantModels(db, "tenant1")

# Offboarding Tenants

When a tenant is removed from the system, the tenant-specific schema and associated tables
should be cleaned up using [DB.OffboardTenant].

	import (
		"context"
		multitenancy "github.com/bartventer/gorm-multitenancy/v8"
		"github.com/bartventer/gorm-multitenancy/postgres/v8"
	)

	func main() {
		db, err := multitenancy.Open(postgres.Open(dsn))
		if err != nil {...}
		db.RegisterModels(ctx, &Tenant{}, &Book{})
		db.MigrateSharedModels(ctx)
		// Assuming we have a tenant with schema name 'tenant1'
		db.MigrateTenantModels(ctx, "tenant1")
		db.OffboardTenant(ctx, "tenant1") // Drop the tenant schema and associated tables
	}

Postgres Adapter:

Use [postgres.DropSchemaForTenant] to offboard a tenant.

	import "github.com/bartventer/gorm-multitenancy/postgres/v8"

	postgres.DropSchemaForTenant(db, "tenant1")

MySQL Adapter:

Use [mysql.DropDatabaseForTenant] to offboard a tenant.

	import "github.com/bartventer/gorm-multitenancy/mysql/v8"

	mysql.DropDatabaseForTenant(db, "tenant1")

# Tenant Context Configuration

[DB.UseTenant] configures the database for operations specific to a tenant,
abstracting database-specific operations for tenant context configuration. This method
returns a reset function to revert the database context and an error if the operation fails.

	import (
		"context"
		multitenancy "github.com/bartventer/gorm-multitenancy/v8"
		"github.com/bartventer/gorm-multitenancy/postgres/v8"
	)

	func main() {
		db, err := multitenancy.Open(postgres.Open(dsn))
		if err != nil {...}
		db.RegisterModels(ctx, &Tenant{}, &Book{})
		db.MigrateSharedModels(ctx)
		// Assuming we have a tenant with schema name 'tenant1'
		reset, err := db.UseTenant(ctx, "tenant1")
		if err != nil {...}
		defer reset() // reset to the default search path
		// ... do operations with the search path set to 'tenant1'
		db.Create(&Book{Title: "The Great Gatsby"})
		db.Find(&Book{})
		db.Delete(&Book{})
	}

Postgres Adapter:

Use [postgres.SetSearchPath] to set the search path for a tenant.

	import "github.com/bartventer/gorm-multitenancy/postgres/v8"

	reset, err := postgres.SetSearchPath(ctx, db, "tenant1")
	if err != nil {...}
	defer reset() // reset to the default search path
	db.Create(&Book{Title: "The Great Gatsby"})

MySQL Adapter:

Use [mysql.UseDatabase] function to set the database for a tenant.

	import "github.com/bartventer/gorm-multitenancy/mysql/v8"

	reset, err := mysql.UseDatabase(ctx, db, "tenant1")
	if err != nil {...}
	defer reset() // reset to the default database
	db.Create(&Book{Title: "The Great Gatsby"})

# Foreign Key Constraints

The framework supports various types of relationships between tables, each with its own set of
considerations. The term "public schema" refers to tables shared across all tenants, while
"tenant-specific tables" are unique to a single tenant.

Between Tables in the Public Schema:
  - No restrictions on foreign key constraints between tables in the public schema.
  - Example: `public.events` can reference `public.locations` without restrictions.

From Public Schema Tables to Tenant-Specific Tables:
  - Tables in the public schema should not have foreign key constraints to tenant-specific tables
    to maintain schema isolation and ensure data integrity across tenants.
  - Example: `public.users` should not reference a tenant-specific `orders` table.

From Tenant-Specific Tables to Public Schema Tables:
  - Tenant-specific tables can have foreign key constraints to tables in the public schema,
    allowing tenant-specific data to reference shared resources or configurations.
  - Example: A tenant-specific `invoices` table can reference `public.payment_methods`.

Between Tenant-Specific Tables:
  - Tenant-specific tables can have foreign key constraints to other tenant-specific tables within
    the same tenant schema, ensuring all related data is encapsulated within a single tenant's
    schema.
  - Example: Within a tenant's schema, a `projects` table can reference an `employees` table.

# Example

	package main

	import (
		"context"

		"github.com/bartventer/gorm-multitenancy/postgres/v8"
		multitenancy "github.com/bartventer/gorm-multitenancy/v8"
	)

	type Tenant struct{ multitenancy.TenantModel }

	func (Tenant) TableName() string   { return "public.tenants" }
	func (Tenant) IsSharedModel() bool { return true }

	type Book struct {
		gorm.Model
		Title        string
		TenantSchema string
		Tenant       Tenant `gorm:"foreignKey:TenantSchema;references:SchemaName"`
	}

	func (Book) TableName() string   { return "books" }
	func (Book) IsSharedModel() bool { return false }

	func main() {
		// Open a new PostgreSQL connection as usual
		dsn := "postgres://user:password@localhost:5432/dbname?sslmode=disable"
		db, err := multitenancy.Open(postgres.Open(dsn))
		if err != nil {
			panic(err)
		}

		ctx := context.Background()
		if err := db.RegisterModels(ctx, &Tenant{}, &Book{}); err != nil {
			panic(err)
		}

		if err := db.MigrateSharedModels(ctx); err != nil {
			panic(err)
		}

		// Create and manage tenants as needed
		tenant := &Tenant{
			TenantModel: multitenancy.TenantModel{
				DomainURL:  "tenant1.example.com",
				SchemaName: "tenant1",
			},
		}
		// Create a tenant in the default public/shared schema
		if err := db.Create(tenant).Error; err != nil {
			panic(err)
		}

		// Migrate models under the tenant schema
		if err := db.MigrateTenantModels(ctx, tenant.SchemaName); err != nil {
			panic(err)
		}

		// Create a book under the tenant schema
		if err := createBookHandler(ctx, db, "The Great Gatsby", tenant.SchemaName); err != nil {
			panic(err)
		}

		// Drop the tenant schema and associated tables
		if err := db.OffboardTenant(ctx, tenant.SchemaName); err != nil {
			panic(err)
		}
	}

	func createBookHandler(ctx context.Context, tx *multitenancy.DB, title, tenantID string) error {
		// Set the tenant context for the current operation(s)
		reset, err := tx.UseTenant(ctx, tenantID)
		if err != nil {
			return err
		}
		defer reset()

		// Create a book under the tenant schema
		b := &Book{
			Title:        title,
			TenantSchema: tenantID,
		}
		// ... do operations with the search path set to <tenantID>
		return tx.Create(b).Error
	}

See [the example application] for a more comprehensive demonstration of the framework's
capabilities.

# Security Considerations

Always sanitize input to prevent SQL injection vulnerabilities. This framework does not perform any
validation on the database name or schema name parameters. It is the responsibility of the caller to
ensure that these parameters are sanitized. To facilitate this, the framework provides the following
utilities:
  - [pkg/namespace/Validate] to verify tenant names against all supported drivers (MySQL and
    PostgreSQL), ensuring both scheme and database name adhere to expected formats.
  - [middleware/nethttp/ExtractSubdomain] to extract subdomains from HTTP requests, which can be
    used to derive tenant names.

# Design Strategy

For a detailed technical overview of SQL design strategies adopted by the framework, see the
[STRATEGY.md] file.

[postgres.RegisterModels]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/postgres/v8#RegisterModels
[mysql.RegisterModels]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/mysql/v8#RegisterModels
[postgres.MigrateSharedModels]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/postgres/v8#MigrateSharedModels
[mysql.MigrateSharedModels]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/mysql/v8#MigrateSharedModels
[postgres.MigrateTenantModels]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/postgres/v8#MigrateTenantModels
[mysql.MigrateTenantModels]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/mysql/v8#MigrateTenantModels
[postgres.DropSchemaForTenant]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/postgres/v8#OffboardTenant
[mysql.DropDatabaseForTenant]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/mysql/v8#OffboardTenant
[postgres.SetSearchPath]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/postgres/v8#UseTenant
[mysql.UseDatabase]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/mysql/v8#UseDatabase
[the example application]: https://github.com/bartventer/gorm-multitenancy/tree/master/examples/README.md
[pkg/namespace/Validate]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v8/pkg/namespace#Validate
[middleware/nethttp/ExtractSubdomain]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8#ExtractSubdomain
[STRATEGY.md]: https://github.com/bartventer/gorm-multitenancy/tree/master/docs/STRATEGY.md
*/
package multitenancy

import (
	"context"
	"database/sql"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"gorm.io/gorm"
)

type (

	// DB wraps a GORM DB connection, integrating support for multitenancy operations.
	// It provides a unified interface for managing tenant-specific and shared data within
	// a multi-tenant application, leveraging GORM's ORM capabilities for database operations.
	DB struct {
		*gorm.DB
		driver driver.DBFactory
	}
)

// CurrentTenant returns the identifier for the current tenant context or an empty string
// if no context is set.
func (db *DB) CurrentTenant(ctx context.Context) string {
	return db.driver.CurrentTenant(ctx, db.DB)
}

// RegisterModels registers GORM model structs for multitenancy support, preparing models for
// tenant-specific operations.
func (db *DB) RegisterModels(ctx context.Context, models ...driver.TenantTabler) error {
	return db.driver.RegisterModels(ctx, db.DB, models...)
}

// MigrateSharedModels migrates all registered shared/public models.
func (db *DB) MigrateSharedModels(ctx context.Context) error {
	return db.driver.MigrateSharedModels(ctx, db.DB)
}

// MigrateTenantModels migrates all registered tenant-specific models for the specified tenant.
// This method is intended to be used when onboarding a new tenant or updating an existing tenant's
// schema to match the latest model definitions.
func (db *DB) MigrateTenantModels(ctx context.Context, tenantID string) error {
	return db.driver.MigrateTenantModels(ctx, db.DB, tenantID)
}

// OffboardTenant cleans up the database by dropping the tenant-specific schema and associated tables.
// This method is intended to be used after a tenant has been removed.
func (db *DB) OffboardTenant(ctx context.Context, tenantID string) error {
	return db.driver.OffboardTenant(ctx, db.DB, tenantID)
}

// UseTenant configures the database for operations specific to a tenant. A reset function is returned
// to revert the database context to its original state. This method is intended to be used when
// performing operations specific to a tenant, such as creating, updating, or deleting tenant-specific
// data.
func (db *DB) UseTenant(ctx context.Context, tenantID string) (reset func() error, err error) {
	return db.driver.UseTenant(ctx, db.DB, tenantID)
}

// NewDB creates a new [DB] instance using the provided [driver.DBFactory] and [gorm.DB]
// instance. This function is intended for use by custom [Adapter] implementations to
// create new instances of DB with multitenancy support. Not intended for direct
// use in application code.
func NewDB(d driver.DBFactory, tx *gorm.DB) *DB {
	return &DB{
		DB:     tx,
		driver: d,
	}
}

// ======================================================================================
// The below methods have been overridden to return a new DB instance with the updated
// configuration, allowing for method chaining and preserving the multitenancy context.
// ======================================================================================

// Session returns a new copy of the DB, which has a new session with the configuration.
func (db *DB) Session(config *gorm.Session) *DB {
	return NewDB(db.driver, db.DB.Session(config))
}

// Debug starts debug mode.
func (db *DB) Debug() *DB {
	return NewDB(db.driver, db.DB.Debug())
}

// WithContext sets the context for the DB.
func (db *DB) WithContext(ctx context.Context) *DB {
	return NewDB(db.driver, db.DB.WithContext(ctx))
}

// Transaction starts a transaction as a block, returns an error if there's any error
// within the block. If the function passed to tx returns an error, the transaction will
// be rolled back automatically, otherwise, the transaction will be committed.
func (db *DB) Transaction(fc func(tx *DB) error, opts ...*sql.TxOptions) (err error) {
	return db.DB.Transaction(func(tx *gorm.DB) error {
		return fc(NewDB(db.driver, tx))
	}, opts...)
}

// Begin begins a transaction.
func (db *DB) Begin(opts ...*sql.TxOptions) *DB {
	return NewDB(db.driver, db.DB.Begin(opts...))
}
