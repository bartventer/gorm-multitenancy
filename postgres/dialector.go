package postgres

import (
	"fmt"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"

	"github.com/bartventer/gorm-multitenancy/v7/pkg/driver"
	"github.com/bartventer/gorm-multitenancy/v7/pkg/gmterrors"
	"github.com/bartventer/gorm-multitenancy/v7/pkg/logext"
)

type (
	// Config is the PostgreSQL configuration with multitenancy support.
	Config struct {
		postgres.Config
	}

	// Dialector is the PostgreSQL dialector with multitenancy support.
	Dialector struct {
		*postgres.Dialector
		rw       *sync.RWMutex
		registry *driver.ModelRegistry
		logger   *logext.Logger
	}

	// Migrator is the PostgreSQL migrator with multitenancy support.
	Migrator struct {
		postgres.Migrator
		Dialector
	}
)

var _ gorm.Dialector = new(Dialector)

// Open creates a new PostgreSQL dialector with multitenancy support.
func Open(dsn string) gorm.Dialector {
	return &Dialector{
		Dialector: postgres.Open(dsn).(*postgres.Dialector),
		rw:        &sync.RWMutex{},
		registry:  &driver.ModelRegistry{},
		logger:    logext.Default(),
	}
}

// New creates a new PostgreSQL dialector with multitenancy support.
func New(config Config) gorm.Dialector {
	return &Dialector{
		Dialector: postgres.New(config.Config).(*postgres.Dialector),
		rw:        &sync.RWMutex{},
		registry:  &driver.ModelRegistry{},
		logger:    logext.Default(),
	}
}

// Migrator returns a [gorm.Migrator] implementation for the Dialector.
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
		dialector,
	}
}

// RegisterModels registers the given models with the dialector for multitenancy support.
func (dialector *Dialector) RegisterModels(models ...driver.TenantTabler) error {
	registry, err := driver.NewModelRegistry(models...)
	if err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to register models: %w", err))
	}

	dialector.rw.Lock()
	dialector.registry = registry
	dialector.rw.Unlock()
	return nil
}

// RegisterModels registers the given models with the provided [gorm.DB] instance for multitenancy support.
func RegisterModels(db *gorm.DB, models ...driver.TenantTabler) error {
	return db.Dialector.(*Dialector).RegisterModels(models...)
}

// MigratePublicSchema migrates the public schema in the database.
func MigratePublicSchema(db *gorm.DB) error {
	return db.Migrator().(*Migrator).MigrateSharedModels()
}

// MigrateTenantModels creates a new schema for a specific tenant in the PostgreSQL database.
func MigrateTenantModels(db *gorm.DB, schemaName string) error {
	return db.Migrator().(*Migrator).MigrateTenantModels(schemaName)
}

// DropSchemaForTenant drops the schema for a specific tenant in the PostgreSQL database (CASCADE).
func DropSchemaForTenant(db *gorm.DB, schemaName string) error {
	return db.Migrator().(*Migrator).DropSchemaForTenant(schemaName)
}
