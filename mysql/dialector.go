package mysql

import (
	"fmt"
	"sync"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/gmterrors"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/logext"
)

type (
	// Config is the MySQL configuration with multitenancy support.
	Config struct {
		mysql.Config
	}

	// Dialector is the MySQL dialector with multitenancy support.
	Dialector struct {
		*mysql.Dialector
		rw       *sync.RWMutex
		registry *driver.ModelRegistry
		logger   *logext.Logger
	}

	// Migrator is the MySQL migrator with multitenancy support.
	Migrator struct {
		mysql.Migrator
		Dialector
	}
)

var _ gorm.Dialector = new(Dialector)

// Open creates a new MySQL dialector with multitenancy support.
func Open(dsn string) gorm.Dialector {
	return &Dialector{
		Dialector: mysql.Open(dsn).(*mysql.Dialector),
		rw:        &sync.RWMutex{},
		registry:  &driver.ModelRegistry{},
		logger:    logext.Default(),
	}
}

// New creates a new MySQL dialector with multitenancy support.
func New(config Config) gorm.Dialector {
	return &Dialector{
		Dialector: mysql.New(config.Config).(*mysql.Dialector),
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
		mysql.Migrator{
			Migrator: migrator.Migrator{
				Config: migrator.Config{
					DB:        db,
					Dialector: dialector,
				},
			},
			Dialector: *dialector.Dialector,
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

// MigrateSharedModels migrates the public schema in the database.
func MigrateSharedModels(db *gorm.DB) error {
	return db.Migrator().(*Migrator).MigrateSharedModels()
}

// MigrateTenantModels creates a new schema for a specific tenant in the MySQL database.
func MigrateTenantModels(db *gorm.DB, schemaName string) error {
	return db.Migrator().(*Migrator).MigrateTenantModels(schemaName)
}

// DropDatabaseForTenant drops the database for a specific tenant in the MySQL database.
func DropDatabaseForTenant(db *gorm.DB, tenant string) error {
	return db.Migrator().(*Migrator).DropDatabaseForTenant(tenant)
}
