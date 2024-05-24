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