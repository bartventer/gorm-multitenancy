package postgres

import (
	"errors"
	"fmt"

	"github.com/bartventer/gorm-multitenancy/postgres/v8/schema"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/gmterrors"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/migrator"
	"gorm.io/gorm"
)

var _ gorm.Migrator = new(Migrator)

func (m Migrator) AutoMigrate(values ...interface{}) error {
	_, err := migrator.OptionFromDB(m.DB)
	if err != nil {
		return gmterrors.NewWithScheme(DriverName, err)
	}
	return m.Migrator.AutoMigrate(values...)
}

// MigrateTenantModels creates a schema for a specific tenant and migrates the private tables.
func (m Migrator) MigrateTenantModels(tenantID string) error {
	if inTransaction(m.DB) {
		return m.migrateTenantModels(tenantID)
	} else {
		return m.DB.Transaction(func(tx *gorm.DB) error {
			return m.migrateTenantModels(tenantID)
		})
	}
}

func (m Migrator) migrateTenantModels(tenantID string) error {
	m.logger.Printf("⏳ migrating tables for tenant %s", tenantID)
	tenantModels := m.registry.TenantModels
	if len(tenantModels) == 0 {
		return gmterrors.NewWithScheme(DriverName, errors.New("no tenant tables to migrate"))
	}
	tx := m.DB.Session(&gorm.Session{})

	sql := "CREATE SCHEMA IF NOT EXISTS " + tx.Statement.Quote(tenantID)
	if err := tx.Exec(sql).Error; err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to create schema for tenant %s: %w", tenantID, err))
	}

	reset, searchPathErr := schema.SetSearchPath(tx, tenantID)
	if searchPathErr != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to set search path to tenant %s: %w", tenantID, searchPathErr))
	}
	defer reset()

	if err := tx.
		Scopes(migrator.WithOption(migrator.MigratorOption)).
		AutoMigrate(driver.ModelsToInterfaces(tenantModels)...); err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to migrate private tables for tenant %s: %w", tenantID, err))
	}
	m.logger.Printf("✅ private tables migrated for tenant %s", tenantID)

	return nil
}

// MigrateSharedModels migrates the public tables in the database.
func (m Migrator) MigrateSharedModels() error {
	publicModels := m.registry.SharedModels
	if len(publicModels) == 0 {
		return gmterrors.NewWithScheme(DriverName, errors.New("no public tables to migrate"))
	}
	m.logger.Println("⏳ migrating public tables")
	if err := m.DB.
		Scopes(migrator.WithOption(migrator.MigratorOption)).
		AutoMigrate(driver.ModelsToInterfaces(publicModels)...); err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to migrate public tables: %w", err))
	}
	m.logger.Printf("✅ public tables migrated for all tenants")
	return nil
}

// DropSchemaForTenant drops the schema for a specific tenant.
func (m Migrator) DropSchemaForTenant(tenant string) error {
	if inTransaction(m.DB) {
		return m.dropSchemaForTenant(tenant)
	} else {
		return m.DB.Transaction(func(tx *gorm.DB) error {
			return m.dropSchemaForTenant(tenant)
		})
	}
}

func (m Migrator) dropSchemaForTenant(tenant string) error {
	tx := m.DB.Session(&gorm.Session{})
	m.logger.Printf("⏳ dropping schema for tenant %s", tenant)
	sql := "DROP SCHEMA IF EXISTS " + tx.Statement.Quote(tenant) + " CASCADE"
	if err := tx.Exec(sql).Error; err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to drop schema for tenant %s: %w", tenant, err))
	}
	m.logger.Printf("✅ schema dropped for tenant %s", tenant)
	return nil
}

func inTransaction(db *gorm.DB) bool {
	committer, ok := db.Statement.ConnPool.(gorm.TxCommitter)
	return ok && committer != nil
}
