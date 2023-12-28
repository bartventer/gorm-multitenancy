# gorm-multitenancy

[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gorm-multitenancy.svg)](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy)
[![Go Report Card](https://goreportcard.com/badge/github.com/bartventer/gorm-multitenancy)](https://goreportcard.com/report/github.com/bartventer/gorm-multitenancy)
[![Coverage Status](https://coveralls.io/repos/github/bartventer/gorm-multitenancy/badge.svg?branch=master)](https://coveralls.io/github/bartventer/gorm-multitenancy?branch=master)
[![Build](https://github.com/bartventer/gorm-multitenancy/actions/workflows/go.yml/badge.svg)](https://github.com/bartventer/gorm-multitenancy/actions/workflows/go.yml)
[![License](https://img.shields.io/github/license/bartventer/gorm-multitenancy.svg)](LICENSE)

There are three common approaches to multitenancy in a database:
- Shared database, shared schema
- Shared database, separate schemas
- Separate databases

This package implements the shared database, separate schemas approach. It uses the [gorm](https://gorm.io/) ORM to manage the database and provides custom drivers to support multitenancy. It also provides HTTP middleware to retrieve the tenant from the request and set the tenant in context.

## Database compatibility
Current supported databases are listed below. Pull requests for other drivers are welcome.
- [PostgreSQL](https://www.postgresql.org/)

## Router compatibility
Current supported routers are listed below. Pull requests for other routers are welcome.
- [echo](https://echo.labstack.com/docs)
- [net/http](https://golang.org/pkg/net/http/)

## Installation

```bash
go get -u github.com/bartventer/gorm-multitenancy
```

## Usage

### PostgreSQL driver
For a complete example, refer to the [examples](#examples) section.
```go

import (
    "gorm.io/gorm"
    "github.com/bartventer/gorm-multitenancy/drivers/postgres"
)

// For models that are tenant specific, ensure that TenantTabler is implemented
// This classifies the model as a tenant specific model when performing subsequent migrations

// Tenant is a public model
type Tenant struct {
    gorm.Model // Embed the gorm.Model
    postgres.TenantModel // Embed the TenantModel
}

// Implement the gorm.Tabler interface
func (t *Tenant) TableName() string {return "tenants"}

// Book is a tenant specific model
type Book struct {
    gorm.Model // Embed the gorm.Model
    Title string
}

// Implement the gorm.Tabler interface
func (b *Book) TableName() string {return "books"}

// Implement the TenantTabler interface
func (b *Book) IsTenantTable() bool {return true}

func main(){
    // Create the database connection
    db, err := gorm.Open(postgres.New(postgres.Config{
        DSN:                  "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable",
    }), &gorm.Config{})
    if err != nil {
        panic(err)
    }

    // Register the models
    // Models are categorized as either public or tenant specific, which allow for simpler migrations
    postgres.RegisterModels(
        db,        // Database connection
        // Public models (does not implement TenantTabler or implements TenantTabler with IsTenantTable() returning false)
        &Tenant{},  
        // Tenant specific model (implements TenantTabler)
        &Book{},
        )

    // Migrate the database
    // Calling AutoMigrate won't work, you must either call MigratePublicSchema or CreateSchemaForTenant
    // MigratePublicSchema will create the public schema and migrate all public models
    // CreateSchemaForTenant will create the schema for the tenant and migrate all tenant specific models

    // Migrate the public schema (migrates all public models)
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

    // Migrate the tenant schema
    // This will create the schema and migrate all tenant specific models
    if err := postgres.CreateSchemaForTenant(db, tenant); err != nil {
        panic(err)
    }

    // Operations on tenant specific schemas (e.g. CRUD operations on books)
    // Refer to Examples section for more details on how to use the middleware

    // Drop the tenant schema
    // This will drop the schema and all tables in the schema
    if err := postgres.DropSchemaForTenant(db, tenant); err != nil {
        panic(err)
    }

    // ... other operations
}
```

## Examples

- [PostgreSQL with echo](https://github.com/bartventer/gorm-multitenancy/tree/master/internal/examples/echo)
- [PostgreSQL with net/http](https://github.com/bartventer/gorm-multitenancy/tree/master/internal/examples/nethttp)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Contributing

All contributions are welcome! Open a pull request to request a feature or submit a bug report.