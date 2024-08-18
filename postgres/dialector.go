package postgres

import (
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/migrator"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/backoff"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/gmterrors"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/logext"
)

type (
	// Options provides configuration options with multitenancy support.
	// By default, retry is enabled. To disable retry, set DisableRetry to true.
	// Note that the retry logic is only applied to migrations.
	Options struct {
		DisableRetry bool            `json:"disable_retry" mapstructure:"disable_retry"` // Whether to disable retry.
		Retry        backoff.Options `json:",inline"       mapstructure:",squash"`       // Retry options.
	}

	// Option is a function that modifies an [Options] instance.
	Option func(*Options)

	// Config provides configuration with multitenancy support.
	Config struct {
		postgres.Config
	}

	// Dialector provides a dialector with multitenancy support.
	Dialector struct {
		*postgres.Dialector
		registry *driver.ModelRegistry
		logger   *logext.Logger
		options  *Options
	}

	// Migrator provides a migrator with multitenancy support.
	Migrator struct {
		postgres.Migrator
		Dialector
	}
)

func (o *Options) apply(opts ...Option) {
	for _, opt := range opts {
		opt(o)
	}

	if !o.DisableRetry {
		o.Retry.MaxRetries = max(o.Retry.MaxRetries, 6)
		o.Retry.Interval = max(o.Retry.Interval, time.Second*2)
		o.Retry.MaxInterval = max(o.Retry.MaxInterval, time.Second*30)
	}
}

var _ gorm.Dialector = new(Dialector)

// Open creates a new PostgreSQL dialector with multitenancy support.
func Open(dsn string) gorm.Dialector {
	options, err := driver.ParseDSNQueryParams[Options](dsn)
	if err != nil {
		panic(fmt.Errorf("failed to parse DSN query parameters: %w", err))
	}
	options.apply()
	return &Dialector{
		Dialector: postgres.Open(dsn).(*postgres.Dialector),
		registry:  &driver.ModelRegistry{},
		logger:    logext.Default(),
		options:   &options,
	}
}

// New creates a new PostgreSQL dialector with multitenancy support.
func New(config Config, opts ...Option) gorm.Dialector {
	options := &Options{}
	options.apply(opts...)
	return &Dialector{
		Dialector: postgres.New(config.Config).(*postgres.Dialector),
		registry:  &driver.ModelRegistry{},
		logger:    logext.Default(),
		options:   options,
	}
}

// Migrator returns a [gorm.Migrator] implementation for the Dialector.
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
		dialector,
	}
}

// RegisterModels registers the given models with the dialector for multitenancy support.
func (dialector *Dialector) RegisterModels(models ...driver.TenantTabler) error {
	registry, err := driver.NewModelRegistry(models...)
	if err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to register models: %w", err))
	}
	dialector.registry = registry
	return nil
}

// RegisterModels registers the given models with the provided [gorm.DB] instance for multitenancy support.
// Not safe for concurrent use by multiple goroutines.
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
