// Package migrator provides utilities for database migration management.
package migrator

import (
	"github.com/bartventer/gorm-multitenancy/v8/pkg/driver"
	"gorm.io/gorm"
)

const pkgName = "gorm-multitenancy/mysql/internal/migrator"

type key string

const migratorKey key = pkgName + "/migrator"

type option uint

// Define values for [option].
const (
	DefaultOption option = iota
	MigratorOption
)

// WithOption sets the migration option for the database.
func WithOption(opt option) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Set(string(migratorKey), opt)
	}
}

// OptionFromDB retrieves the migration option from the database.
// If the option is not found or is not of the correct type, it returns [driver.ErrInvalidMigration].
func OptionFromDB(db *gorm.DB) (option, error) {
	o, optFound := db.Get(string(migratorKey))

	if !optFound || o == nil {
		return 0, driver.ErrInvalidMigration
	}

	optVal, ok := o.(option)
	if !ok {
		return 0, driver.ErrInvalidMigration
	}

	switch optVal {
	case DefaultOption, MigratorOption:
		return optVal, nil
	default:
		return 0, driver.ErrInvalidMigration
	}
}
