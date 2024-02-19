package postgres

import (
	"context"
	"errors"
	"fmt"
	"sync"

	multicontext "github.com/bartventer/gorm-multitenancy/v3/tenantcontext"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

const (
	// PublicSchemaName is the name of the public schema
	PublicSchemaName = "public"
)

type multitenancyMigrationOption uint

const (
	// multiMigrationOptionMigratePublicTables migrates the public tables
	multiMigrationOptionMigratePublicTables multitenancyMigrationOption = iota + 1
	// multiMigrationOptionMigrateTenantTables migrates the tenant tables
	multiMigrationOptionMigrateTenantTables
)

type multitenancyConfig struct {
	publicModels []interface{} // public models are tables that are shared between tenants
	tenantModels []interface{} // tenant models are tables that are private to a tenant
	models       []interface{} // models are all models
}

// Migrator is the struct that implements the Migratorer interface
type Migrator struct {
	postgres.Migrator                 // gorm postgres migrator
	*multitenancyConfig               // config to store the models
	rw                  *sync.RWMutex // mutex to prevent concurrent access to the config
}

var _ MultitenancyMigrator = (*Migrator)(nil)

// CreateSchemaForTenant creates the schema for the tenant and migrates the private tables
func (m *Migrator) CreateSchemaForTenant(tenant string) error {
	m.rw.RLock()
	defer m.rw.RUnlock()

	return m.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		// create schema for tenant
		if err = tx.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", tenant)).Error; err != nil {
			return fmt.Errorf("failed to create schema for tenant %s: %w", tenant, err)
		}
		// set search path to tenant
		if err = setSearchPath(tx, tenant); err != nil {
			return err
		}
		defer func() {
			setSearchPath(tx, PublicSchemaName)
		}()

		// migrate private tables
		if len(m.tenantModels) == 0 {
			return errors.New("no private tables to migrate")
		}
		if tx.Config.Logger != nil {
			tx.Config.Logger.Info(context.Background(), "[multitenancy] ⏳ migrating private tables for tenant '%s'...\n", tenant)
			defer func() {
				if err != nil {
					tx.Config.Logger.Error(context.Background(), "[multitenancy] failed to migrate private tables for tenant '%s': %v\n", tenant, err)
				} else {
					tx.Config.Logger.Info(context.Background(), "[multitenancy] ✅ private tables migrated for tenant 'ss'\n", tenant)
				}
			}()
		}
		if err = tx.
			Scopes(withMigrationOption(multiMigrationOptionMigrateTenantTables)).
			AutoMigrate(m.tenantModels...); err != nil {
			return err
		}

		return nil
	})
}

// MigratePublicSchema migrates the public tables
func (m *Migrator) MigratePublicSchema() error {
	m.rw.RLock()
	defer m.rw.RUnlock()

	if len(m.publicModels) == 0 {
		return errors.New("no public tables to migrate")
	}
	var err error
	if m.DB.Config.Logger != nil {
		m.DB.Config.Logger.Info(context.Background(), "[multitenancy] ⏳ migrating public tables...\n")
		defer func() {
			if err != nil {
				m.DB.Config.Logger.Error(context.Background(), "[multitenancy] failed to migrate public tables: %v\n", err)
			} else {
				m.DB.Config.Logger.Info(context.Background(), "[multitenancy] ✅ public tables migrated\n")
			}
		}()
	}
	if err = m.DB.
		Scopes(withMigrationOption(multiMigrationOptionMigratePublicTables)).
		AutoMigrate(m.publicModels...); err != nil {
		return err
	}
	return nil
}

// AutoMigrate migrates the tables based on the migration options.
func (m Migrator) AutoMigrate(values ...interface{}) error {
	v, ok := m.DB.Get(multicontext.MigrationOptions.String())
	if !ok {
		return errors.New("no migration options found")
	}
	mo, ok := v.(multitenancyMigrationOption)
	if !ok {
		return errors.New("invalid migration options found")
	}
	switch mo {
	case multiMigrationOptionMigratePublicTables, multiMigrationOptionMigrateTenantTables:
		return m.Migrator.AutoMigrate(values...)
	default:
		return errors.New("invalid migration options found")
	}
}

// DropSchemaForTenant drops the schema for the tenant (CASCADING tables)
func (m *Migrator) DropSchemaForTenant(tenant string) error {
	return m.DB.Transaction(func(tx *gorm.DB) error {
		var err error
		if tx.Config.Logger != nil {
			tx.Config.Logger.Info(context.Background(), "[multitenancy] ⏳ dropping schema for tenant `%s`...\n", tenant)
			defer func() {
				if err != nil {
					tx.Config.Logger.Error(context.Background(), "[multitenancy] failed to drop schema for tenant `%s`: %v\n", tenant, err)
				} else {
					tx.Config.Logger.Info(context.Background(), "[multitenancy] ✅ schema dropped for tenant `%s`\n", tenant)
				}
			}()
		}
		if err = tx.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", tenant)).Error; err != nil {
			return fmt.Errorf("[multitenancy] failed to drop schema for tenant %s: %w", tenant, err)
		}
		return nil
	})
}

func withMigrationOption(option multitenancyMigrationOption) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Set(multicontext.MigrationOptions.String(), option)
	}
}

func setSearchPath(db *gorm.DB, schema string) error {
	return db.Exec(fmt.Sprintf("SET search_path TO %s", schema)).Error
}
