package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"

	pgschema "github.com/bartventer/gorm-multitenancy/postgres/v7/schema"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
		m.DB.Logger.Info(context.Background(), formatLogMessage("⏳ migrating private tables", tenant, nil))
		defer func() {
			if err != nil {
				m.DB.Logger.Error(context.Background(), formatLogMessage("failed to migrate private tables", tenant, err))
			} else {
				m.DB.Logger.Info(context.Background(), formatLogMessage("✅ private tables migrated", tenant, nil))
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
	m.DB.Logger.Info(context.Background(), formatLogMessage("⏳ migrating public tables", "all tenants", nil))
	defer func() {
		if err != nil {
			m.DB.Logger.Error(context.Background(), formatLogMessage("failed to migrate public tables", "all tenants", err))
		} else {
			m.DB.Logger.Info(context.Background(), formatLogMessage("✅ public tables migrated", "all tenants", nil))
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
		m.DB.Logger.Info(context.Background(), formatLogMessage("⏳ dropping schema", tenant, nil))
		defer func() {
			if err != nil {
				m.DB.Logger.Error(context.Background(), formatLogMessage("failed to drop schema", tenant, err))
			} else {
				m.DB.Logger.Info(context.Background(), formatLogMessage("✅ schema dropped", tenant, nil))
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

// formatLogMessage formats a log message for a specific action.
func formatLogMessage(action, tenant string, err error) string {
	if err != nil {
		return fmt.Sprintf("[multitenancy/postgres] %s for tenant `%s`: %v\n", action, tenant, err)
	}
	return fmt.Sprintf("[multitenancy/postgres] %s for tenant `%s`\n", action, tenant)
}
