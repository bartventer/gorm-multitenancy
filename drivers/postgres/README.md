# PostgreSQL driver

[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gorm-multitenancy.svg)](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/drivers/postgres/v7)

The PostgreSQL driver provides multitenancy support for PostgreSQL databases using `GORM`.

> This package is a thin wrapper around the `GORM` [postgres driver](https://github.com/go-gorm/postgres), enhancing it with multitenancy support.

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

All tenant-specific models must implement the [TenantTabler](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v7/#TenantTabler) interface, which classifies the model as a tenant-specific model. The `TenantTabler` interface is used to determine which models to migrate when calling `MigratePublicSchema` or `CreateSchemaForTenant`.

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

Call [`RegisterModels`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/drivers/postgres/v7#RegisterModels) after initializing the database to register all models.

```go
import (
    "gorm.io/gorm"
    "github.com/bartventer/gorm-multitenancy/drivers/postgres/v7"
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

Alternatively, you can pass the models as variadic arguments to [`postgres.New`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/drivers/postgres/v7#New) when creating the dialect.

```go
import (
    "gorm.io/gorm"
    "github.com/bartventer/gorm-multitenancy/drivers/postgres/v7"
)

db, err := gorm.Open(postgres.New(postgres.Config{
        DSN: "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable",
    },&Tenant{}, &Book{}), &gorm.Config{})
if err != nil {
    panic(err)
}
```

Or pass the models as variadic arguments to [`postgres.Open`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/drivers/postgres/v7#Open) when creating the dialect.

```go
import (
    "gorm.io/gorm"
    "github.com/bartventer/gorm-multitenancy/drivers/postgres/v7"
)

db, err := gorm.Open(postgres.Open(dsn, &Tenant{}, &Book{}), &gorm.Config{})
if err != nil {
    panic(err)
}
```

### Migrations

After all models have been [registered](#model-registration), we can perform table migrations.

#### Public Tables

Call [`MigratePublicSchema`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/drivers/postgres/v7#MigratePublicSchema) to create the public schema and migrate all public models.

```go
import (
    "github.com/bartventer/gorm-multitenancy/drivers/postgres/v7"
)

err := postgres.MigratePublicSchema(db)
```

#### Tenant Tables

Call [`CreateSchemaForTenant`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/drivers/postgres/v7#CreateSchemaForTenant) to create the schema for a tenant and migrate all tenant-specific models.

```go
import (
    "github.com/bartventer/gorm-multitenancy/drivers/postgres/v7"
)

err := postgres.CreateSchemaForTenant(db, tenantSchemaName)
```

### Dropping Tenant Schemas

Call [`DropSchemaForTenant`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/drivers/postgres/v7#DropSchemaForTenant) to drop the schema and cascade all schema tables.

```go
import (
    "github.com/bartventer/gorm-multitenancy/drivers/postgres/v7"
)

err := postgres.DropSchemaForTenant(db, tenantSchemaName)
```

### Foreign Key Constraints

Conforming to the above conventions, foreign key constraints between public and tenant-specific models can be created just as if you were using a shared database and schema.

You can embed the [postgres.TenantModel](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/drivers/postgres/v7#TenantModel) struct in your tenant model to add the necessary fields for the tenant model.

Then create a foreign key constraint between the public and tenant-specific models using the `SchemaName` field as the foreign key.

```go
import (
    "github.com/bartventer/gorm-multitenancy/drivers/postgres/v7"
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

Use the [`WithTenantSchema`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/drivers/postgres/v7/scopes#WithTenantSchema) scope function when you want to perform operations on a tenant specific table, which may include foreign key constraints to a public schema table(s).

```go
db.Scopes(WithTenantSchema(tenantID)).Find(&Book{})
```

#### `SetSearchPath`

Use the [`SetSearchPath`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/drivers/postgres/v7/schema#SetSearchPath) function when the tenant schema table has foreign key constraints you want to access belonging to other tables in the same tenant schema (and or foreign key relations to public tables).

```go
import (
    pgschema "github.com/bartventer/gorm-multitenancy/drivers/postgres/v7/schema"
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

#### Benchmarks

The benchmarks were run with the following configuration:

- goos: linux
- goarch: amd64
- pkg: github.com/bartventer/gorm-multitenancy/drivers/postgres/v7/schema
- cpu: AMD EPYC 7763 64-Core Processor                
- go version: 1.22.4
- date: 2024-06-18

> The benchmark results were generated during a GitHub Actions workflow run on a Linux runner ([view workflow](https://github.com/bartventer/gorm-multitenancy/actions/runs/9565315909)).

The following table shows the benchmark results, obtained by running:
```bash
go test -bench=^BenchmarkScopingQueries$ -run=^$ -benchmem -benchtime=2s github.com/bartventer/gorm-multitenancy/drivers/postgres/v7/schema
```
> ns/op: nanoseconds per operation (*lower is better*), B/op: bytes allocated per operation (*lower is better*), allocs/op: allocations per operation (*lower is better*)

| Benchmark | ns/op | B/op | allocs/op |
|-----------|-------|------|-----------|
| BenchmarkScopingQueries/Create/SetSearchPath-4 | 1158551 | 17548 | 224 |
| BenchmarkScopingQueries/Create/WithTenantSchema-4 | 892393 | 16236 | 209 |
| BenchmarkScopingQueries/Find/SetSearchPath-4 | 938606 | 6375 | 102 |
| BenchmarkScopingQueries/Find/WithTenantSchema-4 | 667922 | 5076 | 87 |
| BenchmarkScopingQueries/Update/SetSearchPath-4 | 1649238 | 14719 | 209 |
| BenchmarkScopingQueries/Update/WithTenantSchema-4 | 1362411 | 13656 | 205 |
| BenchmarkScopingQueries/Delete/SetSearchPath-4 | 1655629 | 12240 | 190 |
| BenchmarkScopingQueries/Delete/WithTenantSchema-4 | 1438684 | 11299 | 185 |


## Basic Example

Here's a simplified example of how to use the `gorm-multitenancy` package with the PostgreSQL driver:

```go
package main

import (
    "gorm.io/gorm"
    "github.com/bartventer/gorm-multitenancy/drivers/postgres/v7"
    "github.com/bartventer/gorm-multitenancy/drivers/postgres/v7/scopes"
)

// Tenant is a public model
type Tenant struct {
    gorm.Model
    postgres.TenantModel // Embed the TenantModel
}

// Implement the gorm.Tabler interface
func (Tenant) TableName() string {return "public.tenants"} // Note the public. prefix

// Book is a tenant specific model
type Book struct {
    gorm.Model
    Title        string
    TenantSchema string `gorm:"column:tenant_schema"`
    Tenant       Tenant `gorm:"foreignKey:TenantSchema;references:SchemaName"`
}

// Implement the gorm.Tabler interface
func (Book) TableName() string {return "books"} // Note the lack of prefix

// Implement the TenantTabler interface
func (Book) IsTenantTable() bool {return true} // This classifies the model as a tenant specific model

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


---

_Note: This file was auto-generated by the [update_readme.sh](https://github.com/bartventer/gorm-multitenancy/blob/master/scripts/update_readme.sh) script. Do not edit this file directly._
