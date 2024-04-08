package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"

	pgschema "github.com/bartventer/gorm-multitenancy/v5/schema/postgres"
	"github.com/bartventer/gorm-multitenancy/v5/tenantcontext"
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

type multitenancyConfig struct {
	publicModels []interface{} // public models are tables that are shared between tenants
	tenantModels []interface{} // tenant models are tables that are private to a tenant
	models       []interface{} // models are all models
}

// Migrator is the struct that implements the [MultitenancyMigrator] interface.
type Migrator struct {
	postgres.Migrator                 // gorm postgres migrator
	*multitenancyConfig               // config to store the models
	rw                  *sync.RWMutex // mutex to prevent concurrent access to the config
}

var _ MultitenancyMigrator = (*Migrator)(nil)

// CreateSchemaForTenant creates a schema for a specific tenant in the database.
// It first checks if the schema already exists, and if not, it creates the schema.
// Then, it sets the search path to the newly created schema.
// After that, it migrates the private tables for the specified tenant.
// If there are no private tables to migrate, it returns an error.
// The function returns an error if any of the steps fail.
func (m *Migrator) CreateSchemaForTenant(tenant string) error {
	m.rw.RLock()
	defer m.rw.RUnlock()

	return m.DB.Transaction(func(tx *gorm.DB) error {
		// create schema for tenant
		if err := tx.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", tx.Statement.Quote(tenant))).Error; err != nil {
			return fmt.Errorf("failed to create schema for tenant %s: %w", tenant, err)
		}
		// // set search path to tenant
		tx, resetSearchPath := pgschema.SetSearchPath(tx, tenant)
		if tx.Error != nil {
			return fmt.Errorf("failed to set search path to tenant %s: %w", tenant, tx.Error)
		}
		defer func() {
			_ = resetSearchPath()
		}()

		// migrate private tables
		if len(m.tenantModels) == 0 {
			return errors.New("no private tables to migrate")
		}
		var err error
		if logger := tx.Config.Logger; logger != nil {
			logger.Info(context.Background(), "[multitenancy] ⏳ migrating private tables for tenant '%s'...\n", tenant)
			defer func() {
				if err != nil {
					logger.Error(context.Background(), "[multitenancy] failed to migrate private tables for tenant '%s': %v\n", tenant, err)
				} else {
					logger.Info(context.Background(), "[multitenancy] ✅ private tables migrated for tenant 'ss'\n", tenant)
				}
			}()
		}
		if err = tx.
			Scopes(withMigrationOption(migrationOptionTenantTables)).
			AutoMigrate(m.tenantModels...); err != nil {
			return err
		}

		return nil
	})
}

// MigratePublicSchema migrates the public tables in the database.
// It checks if there are any public tables to migrate and then performs the migration.
// If an error occurs during the migration, it logs the error.
// This function returns an error if there are no public tables to migrate or if an error occurs during the migration.
func (m *Migrator) MigratePublicSchema() error {
	m.rw.RLock()
	defer m.rw.RUnlock()

	if len(m.publicModels) == 0 {
		return errors.New("no public tables to migrate")
	}
	var err error
	if logger := m.DB.Config.Logger; logger != nil {
		logger.Info(context.Background(), "[multitenancy] ⏳ migrating public tables...\n")
		defer func() {
			if err != nil {
				logger.Error(context.Background(), "[multitenancy] failed to migrate public tables: %v\n", err)
			} else {
				logger.Info(context.Background(), "[multitenancy] ✅ public tables migrated\n")
			}
		}()
	}
	if err = m.DB.
		Scopes(withMigrationOption(migrationOptionPublicTables)).
		AutoMigrate(m.publicModels...); err != nil {
		return err
	}
	return nil
}

// AutoMigrate migrates the specified values to the database.
// It checks for migration options in the context and performs the migration accordingly.
// If no migration options are found or if the migration options are invalid, an error is returned.
// The supported migration options are migrationOptionPublicTables and migrationOptionTenantTables.
// For any other migration option, an error is returned.
func (m Migrator) AutoMigrate(values ...interface{}) error {
	v, ok := m.DB.Get(tenantcontext.MigrationOptions.String())
	if !ok {
		return errors.New("no migration options found")
	}
	mo, ok := v.(migrationOption)
	if !ok {
		return errors.New("invalid migration options found")
	}
	switch mo {
	case migrationOptionPublicTables, migrationOptionTenantTables:
		return m.Migrator.AutoMigrate(values...)
	default:
		return errors.New("invalid migration options found")
	}
}

// DropSchemaForTenant drops the schema for a specific tenant.
// It executes a transaction and drops the schema using the provided tenant name.
// If an error occurs during the transaction or while dropping the schema, it returns an error.
// Otherwise, it returns nil.
func (m *Migrator) DropSchemaForTenant(tenant string) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		if logger := tx.Config.Logger; logger != nil {
			logger.Info(context.Background(), "[multitenancy] ⏳ dropping schema for tenant `%s`...\n", tenant)
			defer func() {
				if err != nil {
					logger.Error(context.Background(), "[multitenancy] failed to drop schema for tenant `%s`: %v\n", tenant, err)
				} else {
					logger.Info(context.Background(), "[multitenancy] ✅ schema dropped for tenant `%s`\n", tenant)
				}
			}()
		}
		if err = tx.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenant)).Error; err != nil {
			return fmt.Errorf("[multitenancy] failed to drop schema for tenant %s: %w", tenant, err)
		}
		return nil
	})
}

// withMigrationOption is a higher-order function that returns a function which sets the migration option for a GORM database connection.
// The returned function takes a *gorm.DB parameter and returns a modified *gorm.DB with the migration option set.
// The migration option is set using the tenantcontext.MigrationOptions.String() key and the provided option value.
func withMigrationOption(option migrationOption) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Set(tenantcontext.MigrationOptions.String(), option)
	}
}
