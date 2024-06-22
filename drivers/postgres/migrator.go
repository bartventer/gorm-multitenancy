package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"

	pgschema "github.com/bartventer/gorm-multitenancy/drivers/postgres/v7/schema"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	// PublicSchemaName is the name of the public schema.
	PublicSchemaName = "public"
)

type migrationOption uint

const (
	// migrationOptionPublicTables migrates the public tables.
	migrationOptionPublicTables migrationOption = iota + 1
	// migrationOptionTenantTables migrates the tenant tables.
	migrationOptionTenantTables
)

type migratorConfig struct {
	publicModels []interface{} // public models are tables that are shared between tenants
	tenantModels []interface{} // tenant models are tables that are private to a tenant
}

// Migrator is the struct that implements the [MultitenancyMigrator] interface.
type Migrator struct {
	postgres.Migrator               // gorm postgres migrator
	*migratorConfig                 // config to store the models
	rw                *sync.RWMutex // mutex to prevent concurrent access to the config
}

var _ gorm.Migrator = new(Migrator)

var _ MultitenancyMigrator = (*Migrator)(nil)

// CreateSchemaForTenant creates a schema for a specific tenant and migrates the private tables.
func (m *Migrator) CreateSchemaForTenant(tenant string) error {
	m.rw.RLock()
	tenantModels := m.tenantModels
	m.rw.RUnlock()

	if len(tenantModels) == 0 {
		return errors.New("no private tables to migrate")
	}

	return m.DB.Transaction(func(tx *gorm.DB) error {
		// Create schema for tenant if it doesn't exist
		if err := tx.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", tx.Statement.Quote(tenant))).Error; err != nil {
			return fmt.Errorf("failed to create schema for tenant %s: %w", tenant, err)
		}

		// Set search path to tenant
		tx, resetSearchPath := pgschema.SetSearchPath(tx, tenant)
		if tx.Error != nil {
			return fmt.Errorf("failed to set search path to tenant %s: %w", tenant, tx.Error)
		}
		defer func() {
			_ = resetSearchPath()
		}()

		// Migrate private tables
		var err error
		m.logInfof(context.Background(), "⏳ migrating private tables for tenant '%s'...\n", tenant)
		defer func() {
			if err != nil {
				m.logErrorf(context.Background(), "failed to migrate private tables for tenant '%s': %v\n", tenant, err)
			} else {
				m.logInfof(context.Background(), "✅ private tables migrated for tenant '%s'\n", tenant)
			}
		}()
		if err = tx.
			Scopes(withMigrationOption(migrationOptionTenantTables)).
			AutoMigrate(tenantModels...); err != nil {
			return err
		}

		return nil
	})
}

// MigratePublicSchema migrates the public tables in the database.
func (m *Migrator) MigratePublicSchema() error {
	m.rw.RLock()
	publicModels := m.publicModels
	m.rw.RUnlock()

	if len(publicModels) == 0 {
		return errors.New("no public tables to migrate")
	}
	var err error
	m.logInfof(context.Background(), "⏳ migrating public tables...\n")
	defer func() {
		if err != nil {
			m.logErrorf(context.Background(), "failed to migrate public tables: %v\n", err)
		} else {
			m.logInfof(context.Background(), "✅ public tables migrated\n")
		}
	}()
	if err = m.DB.
		Scopes(withMigrationOption(migrationOptionPublicTables)).
		AutoMigrate(publicModels...); err != nil {
		return err
	}
	return nil
}

// AutoMigrate migrates the specified values to the database based on the migration options.
func (m Migrator) AutoMigrate(values ...interface{}) error {
	opt, ok := m.DB.Get(MigrationOptions.String())
	if !ok {
		return errors.New("no migration options found")
	}
	optValue, ok := opt.(migrationOption)
	if !ok {
		return errors.New("invalid migration options found")
	}
	switch optValue {
	case migrationOptionPublicTables, migrationOptionTenantTables:
		return m.Migrator.AutoMigrate(values...)
	default:
		return errors.New("invalid migration options found")
	}
}

// DropSchemaForTenant drops the schema for a specific tenant.
func (m *Migrator) DropSchemaForTenant(tenant string) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		m.logInfof(context.Background(), "⏳ dropping schema for tenant `%s`...\n", tenant)
		defer func() {
			if err != nil {
				m.logErrorf(context.Background(), "failed to drop schema for tenant `%s`: %v\n", tenant, err)
			} else {
				m.logInfof(context.Background(), "✅ schema dropped for tenant `%s`\n", tenant)
			}
		}()
		if err = tx.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenant)).Error; err != nil {
			return fmt.Errorf("failed to drop schema for tenant %s: %w", tenant, err)
		}
		return nil
	})
}

// withMigrationOption sets the migration option for a GORM database connection.
func withMigrationOption(opt migrationOption) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Set(MigrationOptions.String(), opt)
	}
}

func (m *Migrator) logf(ctx context.Context, level logger.LogLevel, format string, args ...interface{}) {
	if log := m.DB.Config.Logger; log != nil {
		format = "[multitenancy] " + format
		switch level {
		case logger.Error:
			log.Error(ctx, format, args...)
		case logger.Warn:
			log.Warn(ctx, format, args...)
		case logger.Info:
			log.Info(ctx, format, args...)
		case logger.Silent:
			// do nothing
		}
	}
}
func (m *Migrator) logErrorf(ctx context.Context, format string, args ...interface{}) {
	m.logf(ctx, logger.Error, format, args...)
}

func (m *Migrator) logInfof(ctx context.Context, format string, args ...interface{}) {
	m.logf(ctx, logger.Info, format, args...)
}
