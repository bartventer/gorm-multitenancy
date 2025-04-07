package postgres

import (
	"errors"
	"fmt"

	"github.com/bartventer/gorm-multitenancy/postgres/v8/internal/locking"
	"github.com/bartventer/gorm-multitenancy/postgres/v8/schema"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/backoff"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/gmterrors"
	"github.com/bartventer/gorm-multitenancy/v8/pkg/migrator"
	"gorm.io/gorm"
)

func (m Migrator) retry(fn func() error) error {
	if !m.options.DisableRetry {
		return backoff.Retry(fn, func(o *backoff.Options) {
			*o = m.options.Retry
		})
	}
	return fn()
}

func (m Migrator) acquireXact(tx *gorm.DB, lockKey string) error {
	return locking.AcquireXact(tx, lockKey, locking.WithRetry(&m.options.Retry))
}

func (m Migrator) AutoMigrate(values ...interface{}) error {
	_, err := migrator.OptionFromDB(m.DB)
	if err != nil {
		return gmterrors.NewWithScheme(DriverName, err)
	}
	return m.retry(func() error {
		return m.Migrator.AutoMigrate(values...)
	})
}

// MigrateTenantModels creates a schema for a specific tenant and migrates the private tables.
func (m Migrator) MigrateTenantModels(tenantID string) error {
	m.logger.Printf("⏳ migrating tables for tenant %s", tenantID)

	tenantModels := m.registry.TenantModels
	if len(tenantModels) == 0 {
		return gmterrors.NewWithScheme(DriverName, errors.New("no tenant tables to migrate"))
	}

	sqlstr := "CREATE SCHEMA IF NOT EXISTS " + m.DB.Statement.Quote(tenantID)
	if err := m.DB.Exec(sqlstr).Error; err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to create schema for tenant %s: %w", tenantID, err))
	}

	err := m.DB.Transaction(func(tx *gorm.DB) error {
		err := m.acquireXact(tx, tenantID)
		if err != nil {
			return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to acquire advisory lock for tenant %s: %w", tenantID, err))
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
	})
	if err != nil {
		return err
	}
	return nil
}

// MigrateSharedModels migrates the public tables in the database.
func (m Migrator) MigrateSharedModels() error {
	m.logger.Println("⏳ migrating public tables")

	publicModels := m.registry.SharedModels
	if len(publicModels) == 0 {
		return gmterrors.NewWithScheme(DriverName, errors.New("no public tables to migrate"))
	}

	tx := m.DB.Begin()
	defer func() {
		if tx.Error == nil {
			tx.Commit()
			m.logger.Printf("✅ public tables migrated for all tenants")
		} else {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to begin transaction: %w", err))
	}

	err := m.acquireXact(tx, driver.PublicSchemaName())
	if err != nil {
		tx.Rollback()
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to acquire advisory lock: %w", err))
	}

	if err := tx.
		Scopes(migrator.WithOption(migrator.MigratorOption)).
		AutoMigrate(driver.ModelsToInterfaces(publicModels)...); err != nil {
		tx.Rollback()
		return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to migrate public tables: %w", err))
	}

	return tx.Commit().Error
}

// DropSchemaForTenant drops the schema for a specific tenant.
func (m Migrator) DropSchemaForTenant(tenant string) error {
	m.logger.Printf("⏳ dropping schema for tenant %s", tenant)

	err := m.retry(func() error {
		sqlstr := "DROP SCHEMA IF EXISTS " + m.DB.Statement.Quote(tenant) + " CASCADE"
		if err := m.DB.Exec(sqlstr).Error; err != nil {
			return gmterrors.NewWithScheme(DriverName, fmt.Errorf("failed to drop schema for tenant %s: %w", tenant, err))
		}
		m.logger.Printf("✅ schema dropped for tenant %s", tenant)
		return nil
	})

	return err
}
