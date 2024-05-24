# PostgreSQL driver

[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gorm-multitenancy.svg)](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v6/drivers/postgres)

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