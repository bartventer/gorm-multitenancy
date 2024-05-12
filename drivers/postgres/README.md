# PostgreSQL driver

[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gorm-multitenancy.svg)](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v6/drivers/postgres)

The PostgreSQL driver provides multitenancy support for PostgreSQL databases using the `gorm` ORM.

> [!NOTE]
> The driver is a thin wrapper around the [gorm.io/driver/postgres](https://pkg.go.dev/gorm.io/driver/postgres) driver, enhancing it with multitenancy support while preserving all the functionalities of the original driver.

## Conventions

### TableName

The driver uses the `public` schema for public models and the tenant-specific schema for tenant-specific models. All models must implement the [gorm.Tabler](https://pkg.go.dev/gorm.io/gorm/schema#Tabler) interface.

#### Public Model

The table name for public models must be prefixed with `public.`.

```go
type Tenant struct {
    ID uint `gorm:"primaryKey"`
    // other fields...
}

func (Tenant) TableName() string {
    return "public.tenants"
}
```

#### Tenant Model

The table name for tenant-specific models must not contain any prefix.

```go
type Book struct {
    ID uint `gorm:"primaryKey"`
    // other fields...
}

func (Book) TableName() string {
    return "books"
}

```

### TenantTabler

All tenant-specific models must implement the [TenantTabler](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v6/#TenantTabler) interface, which classifies the model as a tenant-specific model. The `TenantTabler` interface is used to determine which models to migrate when calling `MigratePublicSchema` or `CreateSchemaForTenant`.

```go
type Book struct {
    ID uint `gorm:"primaryKey"`
    // other fields...
}

func (Book) IsTenantTable() bool {
    return true
}
```

### Model Registration

#### After DB Initialization

Call [`RegisterModels`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v6/drivers/postgres#RegisterModels) after initializing the database to register all models.

```go
import (
    "gorm.io/gorm"
    "github.com/bartventer/gorm-multitenancy/v6/drivers/postgres"
)

db, err := gorm.Open(postgres.New(postgres.Config{
        DSN: "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable",
    }), &gorm.Config{})
if err != nil {
    panic(err)
}
err := postgres.RegisterModels(db, &Tenant{}, &Book{})
```

#### During DB Initialization

Alternatively, you can pass the models as variadic arguments to [`postgres.New`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v6/drivers/postgres#New) when creating the dialect.

```go
import (
    "gorm.io/gorm"
    "github.com/bartventer/gorm-multitenancy/v6/drivers/postgres"
)

db, err := gorm.Open(postgres.New(postgres.Config{
        DSN: "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable",
    },&Tenant{}, &Book{}), &gorm.Config{})
if err != nil {
    panic(err)
}
```

Or pass the models as variadic arguments to [`postgres.Open`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v6/drivers/postgres#Open) when creating the dialect.

```go
import (
    "gorm.io/gorm"
    "github.com/bartventer/gorm-multitenancy/v6/drivers/postgres"
)

db, err := gorm.Open(postgres.Open(dsn, &Tenant{}, &Book{}), &gorm.Config{})
if err != nil {
    panic(err)
}
```

### Migrations

After all models have been [registered](#model-registration), we can perform table migrations.

#### Public Tables

Call [`MigratePublicSchema`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v6/drivers/postgres#MigratePublicSchema) to create the public schema and migrate all public models.

```go
import (
    "github.com/bartventer/gorm-multitenancy/v6/drivers/postgres"
)

err := postgres.MigratePublicSchema(db)
```

#### Tenant Tables

Call [`CreateSchemaForTenant`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v6/drivers/postgres#CreateSchemaForTenant) to create the schema for a tenant and migrate all tenant-specific models.

```go
import (
    "github.com/bartventer/gorm-multitenancy/v6/drivers/postgres"
)

err := postgres.CreateSchemaForTenant(db, tenantSchemaName)
```

### Dropping Tenant Schemas

Call [`DropSchemaForTenant`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v6/drivers/postgres#DropSchemaForTenant) to drop the schema and cascade all schema tables.

```go
import (
    "github.com/bartventer/gorm-multitenancy/v6/drivers/postgres"
)

err := postgres.DropSchemaForTenant(db, tenantSchemaName)
```

### Foreign Key Constraints

Conforming to the above conventions, foreign key constraints between public and tenant-specific models can be created just as if you were using a shared database and schema.

You can embed the [postgres.TenantModel](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v6/drivers/postgres#TenantModel) struct in your tenant model to add the necessary fields for the tenant model.

Then create a foreign key constraint between the public and tenant-specific models using the `SchemaName` field as the foreign key.

```go
import (
    "github.com/bartventer/gorm-multitenancy/v6/drivers/postgres"
    "gorm.io/gorm"
)

type Tenant struct {
    gorm.Model
    postgres.TenantModel
}

func (Tenant) TableName() string {
    return "public.tenants"
}

type Book struct {
    gorm.Model
    TenantSchema string `gorm:"column:tenant_schema"`
    Tenant       Tenant `gorm:"foreignKey:TenantSchema;references:SchemaName"`
}

func (Book) IsTenantTable() bool {
    return true
}

func (Book) TableName() string {
    return "books"
}
```

### Tenant Schema Scopes

#### `WithTenantSchema`

Use the [`WithTenantSchema`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v6/scopes#WithTenantSchema) scope function when you want to perform operations on a tenant specific table, which may include foreign key constraints to a public schema table(s).

```go
db.Scopes(WithTenantSchema(tenantID)).Find(&Book{})
```

#### `SetSearchPath`

Use the [`SetSearchPath`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema#SetSearchPath) function when the tenant schema table has foreign key constraints you want to access belonging to other tables in the same tenant schema (and or foreign key relations to public tables).

```go
import (
    pgschema "github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema"
    "gorm.io/gorm"
)
db, resetSearchPath := pgschema.SetSearchPath(db, tenantSchemaName)
if err := db.Error(); err != nil {
    // handle error
}
defer resetSearchPath()
// No need to use any tenant scopes as the search path has been changed to the tenant's schema
db.Find(&Book{})
```
<!-- from tmp/bench/BenchmarkScopingQueries.md -->
#### Benchmarks

- goos: linux
- goarch: amd64
- pkg: github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/schema
- cpu: Intel(R) Core(TM) i5-7360U CPU @ 2.30GHz
- date: 2024-05-12

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| BenchmarkScopingQueries/Create/WithTenantSchema-4 | 52052621 | 16382 | 207 |
| BenchmarkScopingQueries/Create/SetSearchPath-4 | 156505 | 1672 | 25 |
| BenchmarkScopingQueries/Find/WithTenantSchema-4 | 161192 | 4917 | 86 |
| BenchmarkScopingQueries/Find/SetSearchPath-4 | 328485 | 6375 | 102 |
| BenchmarkScopingQueries/Update/WithTenantSchema-4 | 42719279 | 13418 | 203 |
| BenchmarkScopingQueries/Update/SetSearchPath-4 | 341492 | 6392 | 104 |
| BenchmarkScopingQueries/Delete/WithTenantSchema-4 | 40143705 | 10822 | 178 |
| BenchmarkScopingQueries/Delete/SetSearchPath-4 | 44092282 | 12146 | 187 |
<!-- end from tmp/bench/BenchmarkScopingQueries.md -->

## Basic Example

Here's a simplified example of how to use the `gorm-multitenancy` package with the PostgreSQL driver:

```go
package main

import (
    "gorm.io/gorm"
    "github.com/bartventer/gorm-multitenancy/v6/drivers/postgres"
    "github.com/bartventer/gorm-multitenancy/v6/drivers/postgres/scopes"
)

// Tenant is a public model
type Tenant struct {
    gorm.Model
    postgres.TenantModel // Embed the TenantModel
}

// Implement the gorm.Tabler interface
func (t *Tenant) TableName() string {return "public.tenants"} // Note the public. prefix

// Book is a tenant specific model
type Book struct {
    gorm.Model
    Title        string
    TenantSchema string `gorm:"column:tenant_schema"`
    Tenant       Tenant `gorm:"foreignKey:TenantSchema;references:SchemaName"`
}

// Implement the gorm.Tabler interface
func (b *Book) TableName() string {return "books"} // Note the lack of prefix

// Implement the TenantTabler interface
func (b *Book) IsTenantTable() bool {return true} // This classifies the model as a tenant specific model

func main(){
    // Open a connection to the database
    db, err := gorm.Open(postgres.New(postgres.Config{
        DSN: "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable",
    }), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    // Register models
    if err := postgres.RegisterModels(db, &Tenant{}, &Book{}); err != nil {
        panic(err)
    }

    // Migrate the public schema
    if err := postgres.MigratePublicSchema(db); err != nil {
        panic(err)
    }

    // Create a tenant
    tenant := &Tenant{
        TenantModel: postgres.TenantModel{
            DomainURL: "tenant1.example.com",
            SchemaName: "tenant1",
        },
    }
    if err := db.Create(tenant).Error; err != nil {
        panic(err)
    }

    // Create the schema for the tenant
    if err := postgres.CreateSchemaForTenant(db, tenant.SchemaName); err != nil {
        panic(err)
    }

    // Create a book for the tenant
    b := &Book{
        Title: "Book 1",
        TenantSchema: tenant.SchemaName,
    }
    if err := db.Scopes(scopes.WithTenantSchema(tenant.SchemaName)).Create(b).Error; err != nil {
        panic(err)
    }

    // Drop the schema for the tenant
    if err := postgres.DropSchemaForTenant(db, tenant.SchemaName); err != nil {
        panic(err)
    }
}
```

## Complete Examples

For more detailed examples, including how to use the middleware with different frameworks, please refer to the following:

- [PostgreSQL with echo](../../examples/echo/README.md)
- [PostgreSQL with net/http](../../examples/nethttp/README.md)
