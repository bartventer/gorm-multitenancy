# gorm-multitenancy

[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gorm-multitenancy.svg)](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5)
[![Release](https://img.shields.io/github/release/bartventer/gorm-multitenancy.svg)](https://github.com/bartventer/gorm-multitenancy/releases/latest)
[![Go Report Card](https://goreportcard.com/badge/github.com/bartventer/gorm-multitenancy)](https://goreportcard.com/report/github.com/bartventer/gorm-multitenancy)
[![Coverage Status](https://coveralls.io/repos/github/bartventer/gorm-multitenancy/badge.svg?branch=master)](https://coveralls.io/github/bartventer/gorm-multitenancy?branch=master)
[![Build](https://github.com/bartventer/gorm-multitenancy/actions/workflows/default.yml/badge.svg)](https://github.com/bartventer/gorm-multitenancy/actions/workflows/default.yml)
![GitHub issues](https://img.shields.io/github/issues/bartventer/gorm-multitenancy)
[![License](https://img.shields.io/github/license/bartventer/gorm-multitenancy.svg)](LICENSE)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fbartventer%2Fgorm-multitenancy.svg?type=shield&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fbartventer%2Fgorm-multitenancy?ref=badge_shield&issueType=license)

<p align="center">
  <img src="https://i.imgur.com/bOZB8St.png" title="GORM Multitenancy" alt="GORM Multitenancy">
</p>
<p align="center">
  <sub><small>Photo by <a href="https://github.com/ashleymcnamara">Ashley McNamara</a>, via <a href="https://github.com/ashleymcnamara/gophers">ashleymcnamara/gophers</a> (CC BY-NC-SA 4.0)</small></sub>
</p>

## Multitenancy Approaches

There are three common approaches to multitenancy in a database:

-   Shared database, shared schema
-   Shared database, separate schemas
-   Separate databases

This package implements the shared database, separate schemas approach to multitenancy, providing custom drivers for seamless integration with your existing database setup.

## Features

-   **GORM Integration**: Uses the [gorm](https://gorm.io/) ORM to manage the database, allowing for easy integration with your existing GORM setup.
-   **Custom Database Drivers**: Provides custom drivers to support multitenancy, allowing you to easily swap and change with your existing drivers with minimal initialization reconfiguration.
-   **HTTP Middleware**: Includes middleware for seamless integration with certain routers, enabling the retrieval of the tenant from the request and setting the tenant in context.

## Database compatibility

Current supported databases are listed below. Pull requests for other drivers are welcome.

-   [PostgreSQL](https://www.postgresql.org/)

## Router Integration

This package includes middleware that can be utilized with the routers listed below for seamless integration with the database drivers. While not a requirement, these routers are fully compatible with the provided middleware. Contributions for other routers are welcome.

-   [echo](https://echo.labstack.com/docs)
-   [net/http](https://golang.org/pkg/net/http/)

## Installation

```bash
go get -u github.com/bartventer/gorm-multitenancy/v5
```

## Usage

### PostgreSQL driver

#### Conventions

-   The driver uses the `public` schema for public models and the tenant specific schema for tenant specific models
-   All models must implement the `gorm.Tabler` interface
-   The table name for public models must be prefixed with `public.` (e.g. `public.books`), whereas the table name for tenant specific models must not contain any prefix (e.g. only `books`)
-   All tenant specific models must implement the [TenantTabler](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/#TenantTabler) interface, which classifies the model as a tenant specific model:
    -   The `TenantTabler` interface has a single method `IsTenantTable() bool` which returns `true` if the model is tenant specific and `false` otherwise
    -   The `TenantTabler` interface is used to determine which models to migrate when calling `MigratePublicSchema` or `CreateSchemaForTenant`
-   Models can be registered in two ways:
    -   When creating the dialect, by passing the models as variadic arguments to [`postgres.New`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#New) (e.g. `postgres.New(postgres.Config{...}, &Book{}, &Tenant{})`) or by calling [`postgres.Open`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#Open) (e.g. `postgres.Open("postgres://...", &Book{}, &Tenant{})`)
    -   By calling [`postgres.RegisterModels`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#RegisterModels) (e.g. `postgres.RegisterModels(db, &Book{}, &Tenant{})`)
-   Migrations can be performed in two ways (after registering the models):
    -   By calling [`MigratePublicSchema`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#MigratePublicSchema) to create the public schema and migrate all public models
    -   By calling [`CreateSchemaForTenant`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#CreateSchemaForTenant) to create the schema for the tenant and migrate all tenant specific models
-   To drop a tenant schema, call [`DropSchemaForTenant`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#DropSchemaForTenant); this will drop the schema and all tables in the schema
    -   When creating the dialect, by passing the models as variadic arguments to [`postgres.New`](<(https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#New)>) (e.g. `postgres.New(postgres.Config{...}, &Book{}, &Tenant{})`) or by calling [`postgres.Open`](<(https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#Open)>) (e.g. `postgres.Open("postgres://...", &Book{}, &Tenant{})`)
    -   By calling [`postgres.RegisterModels`](<(https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#RegisterModels)>) (e.g. `postgres.RegisterModels(db, &Book{}, &Tenant{})`)
-   Migrations can be performed in two ways (after registering the models):
    -   By calling [`MigratePublicSchema`](<(https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#MigratePublicSchema)>) to create the public schema and migrate all public models
    -   By calling [`CreateSchemaForTenant`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#CreateSchemaForTenant) to create the schema for the tenant and migrate all tenant specific models
-   To drop a tenant schema, call [`DropSchemaForTenant`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#DropSchemaForTenant); this will drop the schema and cascade all schema tables

#### Foreign Key Constraints

-   Conforming to the [above conventions](#conventions), foreign key constraints between public and tenant specific models can be created just as if you were using approach 1 (shared database, shared schema).
-   The easiest way to get this working is to embed the [postgres.TenantModel](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#TenantModel) struct in your tenant model. This will add the necessary fields for the tenant model (e.g. `DomainURL` and `SchemaName`), you can then create a foreign key constraint between the public and tenant specific models using the `SchemaName` field as the foreign key (e.g. `gorm:"foreignKey:TenantSchema;references:SchemaName"`); off course, you can also create foreign key constraints between any other fields in the models.

#### Operations on Tenant-Specific Models

Outlined below are two approaches to perform operations on tenant specific models. The first approach is for simple operations on tenant specific models, whereas the second approach is for more complex operations on tenant specific models, but does add ~0.200ms overhead per operation.
| Function | Description |
| --- | --- |
| [`WithTenantSchema`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/scopes#WithTenantSchema) | Use this scope function when you want to perform operations on a tenant table, which may include foreign key constraints to a public schema table(s). |
| [`SetSearchPath`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/schema/postgres#SetSearchPath) | Use this function when the tenant schema table has foreign key constraints you want to access belonging to other tables in the same tenant schema (and or foreign key relations to public tables). |

#### Basic example

Here's a simplified example of how to use the `gorm-multitenancy` package with the PostgreSQL driver:

```go

import (
    "gorm.io/gorm"
    "github.com/bartventer/gorm-multitenancy/v5/drivers/postgres"
)

// For models that are tenant specific, ensure that TenantTabler is implemented
// This classifies the model as a tenant specific model when performing subsequent migrations

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
    Title string

    // FK to TenantSchema (same as if you were using approach 1; not realy needed if you use
    // approach 2 as the schema is constrained to the tenant already, but included to show how
    // to create foreign key constraints between public and tenant specific models)
    TenantSchema string `gorm:"column:tenant_schema"`
	Tenant       Tenant `gorm:"foreignKey:TenantSchema;references:SchemaName"`

}

// Implement the gorm.Tabler interface
func (b *Book) TableName() string {return "books"} // Note the lack of prefix

// Implement the TenantTabler interface
func (b *Book) IsTenantTable() bool {return true} // This classifies the model as a tenant specific model

func main(){
    db, err := gorm.Open(postgres.New(postgres.Config{
        DSN: "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable",
    }), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    if err := postgres.RegisterModels(db, &Tenant{}, &Book{}); err != nil {
        panic(err)
    }

    if err := postgres.MigratePublicSchema(db); err != nil {
        panic(err)
    }

    tenant := &Tenant{
        TenantModel: postgres.TenantModel{
            DomainURL: "tenant1.example.com",
            SchemaName: "tenant1",
        },
    }
    if err := db.Create(tenant).Error; err != nil {
        panic(err)
    }

    if err := postgres.CreateSchemaForTenant(db, tenant.SchemaName); err != nil {
        panic(err)
    }

    if err := postgres.DropSchemaForTenant(db, tenant.SchemaName); err != nil {
        panic(err)
    }
}
```

For more detailed examples, including how to use the middleware with different frameworks, please refer to the following:

-   [PostgreSQL with echo](https://github.com/bartventer/gorm-multitenancy/tree/master/examples/echo)
-   [PostgreSQL with net/http](https://github.com/bartventer/gorm-multitenancy/tree/master/examples/nethttp)

## Contributing

All contributions are welcome! Open a pull request to request a feature or submit a bug report.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fbartventer%2Fgorm-multitenancy.svg?type=large&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fbartventer%2Fgorm-multitenancy?ref=badge_large&issueType=license)
