/*
Package postgres provides a [PostgreSQL] driver for [GORM], offering tools to facilitate the construction and management of multi-tenant applications.

Example:

	package main

	import (
		"gorm.io/gorm"
		"github.com/bartventer/gorm-multitenancy/v5/drivers/postgres"
		"github.com/bartventer/gorm-multitenancy/v5/drivers/postgres/scopes"
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

[PostgreSQL]: https://www.postgresql.org
[GORM]: https://gorm.io
*/
package postgres

import (
	"fmt"
	"strings"
	"sync"

	multitenancy "github.com/bartventer/gorm-multitenancy/v5"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
)

type (
	// Config is the configuration for the postgres driver.
	Config = postgres.Config

	// Dialector is the postgres dialector with multitenancy support.
	Dialector struct {
		postgres.Dialector
		*multitenancyConfig
		rw *sync.RWMutex
	}
)

// Check interface.
var _ gorm.Dialector = (*Dialector)(nil)

func initializeDialector(d *Dialector, models ...interface{}) {
	mtc, err := newMultitenancyConfig(models)
	if err != nil {
		panic(err)
	}
	d.multitenancyConfig = mtc
}

// Open opens a connection to a PostgreSQL database using the provided DSN (Data Source Name) and models.
// It returns a gorm.Dialector that can be used to interact with the database.
// The models parameter is optional and can be used to specify the database models that should be registered.
// If an error occurs while creating the multitenancy configuration, it panics.
func Open(dsn string, models ...interface{}) gorm.Dialector {
	d := &Dialector{
		Dialector: *postgres.Open(dsn).(*postgres.Dialector),
		rw:        &sync.RWMutex{},
	}
	initializeDialector(d, models...)
	return d
}

// New creates a new PostgreSQL dialector with multitenancy support.
// It takes a Config struct as the first parameter and variadic models as the second parameter.
// The Config struct contains the necessary configuration for connecting to the PostgreSQL database.
// The models parameter is a list of GORM models that will be used for multitenancy configuration.
// It returns a gorm.Dialector that can be used with GORM.
// If there is an error during the creation of the multitenancy configuration, it will panic.
func New(config Config, models ...interface{}) gorm.Dialector {
	d := &Dialector{
		Dialector: *postgres.New(config).(*postgres.Dialector),
		rw:        &sync.RWMutex{},
	}
	initializeDialector(d, models...)
	return d
}

// Migrator returns a gorm.Migrator implementation for the Dialector.
// It creates a new instance of Migrator with the provided database connection and dialector.
// It also includes a multitenancyConfig that contains information about public models, tenant models, and all models.
// The Migrator is thread-safe and uses a sync.RWMutex for synchronization.
func (dialector Dialector) Migrator(db *gorm.DB) gorm.Migrator {
	dialector.rw.RLock()
	defer dialector.rw.RUnlock()

	return &Migrator{
		postgres.Migrator{
			Migrator: migrator.Migrator{
				Config: migrator.Config{
					DB:                          db,
					Dialector:                   dialector,
					CreateIndexAfterCreateTable: true,
				},
			},
		},
		&multitenancyConfig{
			publicModels: dialector.publicModels,
			tenantModels: dialector.tenantModels,
			models:       dialector.models,
		},
		&sync.RWMutex{},
	}
}

// RegisterModels registers the given models with the provided gorm.DB instance for multitenancy support.
// It initializes the multitenancy configuration for the database dialector.
// The models parameter should be a variadic list of model structs.
// Returns an error if there is a failure in registering the models or initializing the multitenancy configuration.
func RegisterModels(db *gorm.DB, models ...interface{}) error {
	dialector := db.Dialector.(*Dialector)
	mtc, err := newMultitenancyConfig(models)
	if err != nil {
		return fmt.Errorf("failed to register models: %w", err)
	}

	dialector.rw.Lock()
	dialector.multitenancyConfig = mtc
	dialector.rw.Unlock()
	return nil
}

// MigratePublicSchema migrates the public schema in the database.
// It takes a *gorm.DB as input and returns an error if any.
func MigratePublicSchema(db *gorm.DB) error {
	return db.Migrator().(*Migrator).MigratePublicSchema()
}

// CreateSchemaForTenant creates a new schema for a specific tenant in the PostgreSQL database,
// and migrates the private tables for the tenant.
// It takes a gorm.DB instance and the name of the schema as parameters.
// Returns an error if the schema creation fails.
func CreateSchemaForTenant(db *gorm.DB, schemaName string) error {
	return db.Migrator().(*Migrator).CreateSchemaForTenant(schemaName)
}

// DropSchemaForTenant drops the schema for a specific tenant in the PostgreSQL
// database (CASCADING all objects in the schema).
// It takes a *gorm.DB instance and the name of the schema as parameters.
// Returns an error if there was a problem dropping the schema.
func DropSchemaForTenant(db *gorm.DB, schemaName string) error {
	return db.Migrator().(*Migrator).DropSchemaForTenant(schemaName)
}

// newMultitenancyConfig creates a new multitenancy configuration based on the provided models.
// It separates the models into public models and private models based on their table names.
// Public models are those that have a table name starting with the default schema name (PublicSchemaName),
// while private models are those that implement the TenantTabler interface and have a table name without a fullstop.
// If any model does not meet the required criteria, an error message is appended to the errStrings slice.
// If there are any errors, the function returns nil and an error. Otherwise, it returns a new multitenancyConfig
// containing the public models, private models, and all the models.
func newMultitenancyConfig(models []interface{}) (*multitenancyConfig, error) {
	var (
		publicModels  = make([]interface{}, 0, len(models))
		privateModels = make([]interface{}, 0, len(models))
		errStrings    = make([]string, 0)
	)
	for _, m := range models {
		tn, ok := m.(interface{ TableName() string })
		if !ok {
			errStrings = append(errStrings, fmt.Sprintf("model %T does not implement TableName()", m))
			continue
		}
		tt, ok := m.(multitenancy.TenantTabler)
		parts := strings.Split(tn.TableName(), ".")
		if ok && tt.IsTenantTable() {
			// ensure that the private model does not contain a fullstop
			if len(parts) > 1 {
				errStrings = append(errStrings, fmt.Sprintf("invalid table name for model %T labeled as tenant table, table name should not contain a fullstop, got '%s'", m, tn.TableName()))
				continue
			}
			privateModels = append(privateModels, m)
		} else {
			// ensure that the public model starts with the default schema (public.)
			if len(parts) != 2 || parts[0] != PublicSchemaName {
				errStrings = append(errStrings, fmt.Sprintf("invalid table name for model %T labeled as public table, table name should start with '%s.', got '%s'", m, PublicSchemaName, tn.TableName()))
				continue
			}
			publicModels = append(publicModels, m)
		}
	}

	// if there are errors, panic
	if len(errStrings) > 0 {
		return nil, fmt.Errorf("failed to create multitenancy config: %s", strings.Join(errStrings, "; "))
	}

	return &multitenancyConfig{
		publicModels: publicModels,
		tenantModels: privateModels,
		models:       models,
	}, nil
}
