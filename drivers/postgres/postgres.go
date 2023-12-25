package postgres

import (
	multitenancy "github.com/bartventer/gorm-multitenancy"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"
)

type (
	// Config is the configuration for the postgres driver
	Config = postgres.Config

	// Dialector is the postgres dialector with multitenancy support
	Dialector struct {
		postgres.Dialector
		*multitenancyConfig
	}
)

// Check interface
var _ gorm.Dialector = (*Dialector)(nil)

// Open creates a new postgres dialector with multitenancy support
func Open(dsn string, models ...interface{}) gorm.Dialector {
	d := &Dialector{
		Dialector:          *postgres.Open(dsn).(*postgres.Dialector),
		multitenancyConfig: newMultitenancyConfig(models),
	}
	return d
}

// New creates a new postgres dialector with multitenancy support
func New(config Config, models ...interface{}) gorm.Dialector {
	d := &Dialector{
		Dialector: *postgres.New(config).(*postgres.Dialector),
	}
	d.multitenancyConfig = newMultitenancyConfig(models)
	return d
}

// Migrator returns the migrator with multitenancy support
func (dialector Dialector) Migrator(db *gorm.DB) gorm.Migrator {
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
	}
}

// RegisterModels registers the models for multitenancy
func RegisterModels(db *gorm.DB, models ...interface{}) {
	dialector := db.Dialector.(*Dialector)
	dialector.multitenancyConfig = newMultitenancyConfig(models)
}

// MigratePublicSchema migrates the public tables
func MigratePublicSchema(db *gorm.DB) error {
	return db.Migrator().(*Migrator).MigratePublicSchema()
}

// CreateSchemaForTenant creates the schema for the tenant, and migrates the private tables
func CreateSchemaForTenant(db *gorm.DB, schemaName string) error {
	return db.Migrator().(*Migrator).CreateSchemaForTenant(schemaName)
}

// DropSchemaForTenant drops the schema for the tenant (CASCADING tables)
func DropSchemaForTenant(db *gorm.DB, schemaName string) error {
	return db.Migrator().(*Migrator).DropSchemaForTenant(schemaName)
}

// newMultitenancyConfig creates a new multitenancy config
func newMultitenancyConfig(models []interface{}) *multitenancyConfig {
	var (
		publicModels  = make([]interface{}, 0, len(models))
		privateModels = make([]interface{}, 0, len(models))
	)
	for _, m := range models {
		tt, ok := m.(multitenancy.TenantTabler)
		if ok && tt.IsTenantTable() {
			privateModels = append(privateModels, m)
		} else {
			publicModels = append(publicModels, m)
		}
	}
	return &multitenancyConfig{
		publicModels: publicModels,
		tenantModels: privateModels,
		models:       models,
	}
}
