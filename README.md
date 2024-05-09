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

## Table of Contents

- [Introduction](#introduction)
- [Multitenancy Approaches](#multitenancy-approaches)
- [Features](#features)
- [Database compatibility](#database-compatibility)
- [Router Integration](#router-integration)
- [Installation](#installation)
- [Usage](#usage)
    <details>
    <summary><a href="#postgresql-driver">PostgreSQL driver</a></summary>

    <ul>
    <li><a href="#conventions">Conventions</a></li>
      <ul>
      <li><a href="#tablename">TableName</a></li>
        <ul>
        <li><a href="#public-model">Public Model</a></li>
        <li><a href="#tenant-model">Tenant Model</a></li>
        </ul>
      <li><a href="#tenanttabler">TenantTabler</a></li>
      <li><a href="#model-registration">Model Registration</a></li>
        <ul>
        <li><a href="#postgresregistermodels">postgres.RegisterModels</a></li>
        <li><a href="#postgresnew">postgres.New</a></li>
        <li><a href="#postgresopen">postgres.Open</a></li>
        </ul>
      <li><a href="#migrations">Migrations</a></li>
        <ul>
        <li><a href="#public-tables">Public Tables</a></li>
        <li><a href="#tenantschema-tables">Tenant/Schema Tables</a></li>
        </ul>
      <li><a href="#dropping-schemas">Dropping Schemas</a></li>
      <li><a href="#foreign-key-constraints">Foreign Key Constraints</a></li>
      <li><a href="#operations-on-tenant-specific-models">Operations on Tenant-Specific Models</a></li>
        <ul>
        <li><a href="#withtenantschema">WithTenantSchema</a></li>
        <li><a href="#setsearchpath">SetSearchPath</a></li>
        </ul>
      <li><a href="#basic-example">Basic Example</a></li>
      <li><a href="#complete-examples">Complete Examples</a></li>
      </ul>
    </ul>

  </details>
- [Contributing](#contributing)
- [License](#license)

## Introduction

Gorm-multitenancy is a Go package that provides a framework for implementing multitenancy in your applications using GORM.

## Multitenancy Approaches

There are three common approaches to multitenancy in a database:

- Shared database, shared schema
- Shared database, separate schemas
- Separate databases

This package implements the shared database, separate schemas approach to multitenancy, providing custom drivers for seamless integration with your existing database setup.

## Features

- **GORM Integration**: Uses the [gorm](https://gorm.io/) ORM to manage the database, allowing for easy integration with your existing GORM setup.
- **Custom Database Drivers**: Provides custom drivers to support multitenancy, allowing you to easily swap and change with your existing drivers with minimal initialization reconfiguration.
- **HTTP Middleware**: Includes middleware for seamless integration with certain routers, enabling the retrieval of the tenant from the request and setting the tenant in context.

## Database compatibility

Current supported databases are listed below. Pull requests for other drivers are welcome.

- [PostgreSQL](https://www.postgresql.org/)

## Router Integration

This package includes middleware that can be utilized with the routers listed below for seamless integration with the database drivers. While not a requirement, these routers are fully compatible with the provided middleware. Contributions for other routers are welcome.

- [echo](https://echo.labstack.com/docs)
- [net/http](https://golang.org/pkg/net/http/)

## Installation

```bash
go get -u github.com/bartventer/gorm-multitenancy/v5
```

## Usage

### PostgreSQL driver

#### Conventions

##### TableName

The driver uses the `public` schema for public models and the tenant-specific schema for tenant-specific models. All models must implement the [gorm.Tabler](https://pkg.go.dev/gorm.io/gorm/schema#Tabler) interface.

###### Public Model

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

###### Tenant Model

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

##### TenantTabler

All tenant-specific models must implement the [TenantTabler](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/#TenantTabler) interface, which classifies the model as a tenant-specific model. The `TenantTabler` interface is used to determine which models to migrate when calling `MigratePublicSchema` or `CreateSchemaForTenant`.

```go
type Book struct {
    ID uint `gorm:"primaryKey"`
    // other fields...
}

func (Book) IsTenantTable() bool {
    return true
}
```

##### Model Registration

Models can be registered by either calling [`postgres.RegisterModels`](#postgresregistermodels) or when creating the dialect, by passing the models as variadic arguments to [`postgres.New`](#postgresnew) or [`postgres.Open`](#postgresopen).

###### postgres.RegisterModels

```go
db, err := gorm.Open(postgres.New(postgres.Config{
        DSN: "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable",
    }), &gorm.Config{})
postgres.RegisterModels(db, &Tenant{}, &Book{})
```

_Further documentation [here]((https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#RegisterModels))_

###### postgres.New

```go
import (
    "gorm.io/gorm"
    "github.com/bartventer/gorm-multitenancy/v5/drivers/postgres"
)

db, err := gorm.Open(postgres.New(postgres.Config{
        DSN: "host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable",
    },&Tenant{}, &Book{}), &gorm.Config{})
```

_Further documentation [here](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#New)_

###### postgres.Open

```go
import (
    "gorm.io/gorm"
    "github.com/bartventer/gorm-multitenancy/v5/drivers/postgres"
)

db, err := gorm.Open(postgres.Open(dsn, &Tenant{}, &Book{}), &gorm.Config{})
```

_Further documentation [here](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#Open)_

##### Migrations

After all models have been [registered](#model-registration), we can perform table migrations.

###### Public Tables

Call [`MigratePublicSchema`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#MigratePublicSchema) to create the public schema and migrate all public models.

```go
db.MigratePublicSchema()
```

###### Tenant/Schema Tables

Call [`CreateSchemaForTenant`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#CreateSchemaForTenant) to create the schema for a tenant and migrate all tenant-specific models.

```go
db.CreateSchemaForTenant(tenantSchemaName)
```

##### Dropping Schemas

Call [`DropSchemaForTenant`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#DropSchemaForTenant) to drop the schema and cascade all schema tables.

```go
db.DropSchemaForTenant(tenantSchemaName)
```

##### Foreign Key Constraints

Conforming to the above conventions, foreign key constraints between public and tenant-specific models can be created just as if you were using a shared database and schema.

You can embed the [postgres.TenantModel](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/drivers/postgres#TenantModel) struct in your tenant model to add the necessary fields for the tenant model.

Then create a foreign key constraint between the public and tenant-specific models using the `SchemaName` field as the foreign key.

```go
import (
    "github.com/bartventer/gorm-multitenancy/v5/drivers/postgres"
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

##### Operations on Tenant-Specific Models

###### WithTenantSchema

Use the [`WithTenantSchema`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/scopes#WithTenantSchema) scope function when you want to perform operations on a tenant specific table, which may include foreign key constraints to a public schema table(s).

```go
db.Scopes(WithTenantSchema(tenantID)).Find(&Book{})
```

###### SetSearchPath

Use the [`SetSearchPath`](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v5/schema/postgres#SetSearchPath) function when the tenant schema table has foreign key constraints you want to access belonging to other tables in the same tenant schema (and or foreign key relations to public tables). _This is for more complex operations but does add ~0.200ms overhead per operation._

```go
import (
    pgschema "github.com/bartventer/gorm-multitenancy/v5/schema/postgres"
    "gorm.io/gorm"
)
db, resetSearchPath := pgschema.SetSearchPath(db, tenantSchemaName)
if db.Error() != nil {
    // handle error
}
defer resetSearchPath()
// No need to use any tenant scopes as the search path has been changed to the tenant's schema
db.Find(&Book{})
```

#### Basic Example

Here's a simplified example of how to use the `gorm-multitenancy` package with the PostgreSQL driver:

```go

import (
    "gorm.io/gorm"
    "github.com/bartventer/gorm-multitenancy/v5/drivers/postgres"
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

#### Complete Examples

For more detailed examples, including how to use the middleware with different frameworks, please refer to the following:

- [PostgreSQL with echo](examples/echo/README.md)
- [PostgreSQL with net/http](examples/nethttp/README.md)

## Contributing

All contributions are welcome! Open a pull request to request a feature or submit a bug report.

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2Fbartventer%2Fgorm-multitenancy.svg?type=large&issueType=license)](https://app.fossa.com/projects/git%2Bgithub.com%2Fbartventer%2Fgorm-multitenancy?ref=badge_large&issueType=license)
