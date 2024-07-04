package mysql

import (
	"github.com/bartventer/gorm-multitenancy/v7/pkg/driver"
	"gorm.io/gorm"
)

type key string

const migratorKey key = "gorm-multitenancy/mysql:migrator"

type option struct{ name string }

func (o option) String() string {
	return "gorm-multitenancy/mysql/option: " + o.name
}

var migratorOption = option{"migrator"}

// withMigrationOption sets the migration option for the database.
func withMigrationOption(opt option) func(*gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Set(string(migratorKey), opt)
	}
}

// migrationOptionFromDB retrieves the migration option from the database.
func migrationOptionFromDB(db *gorm.DB) (*option, error) {
	o, optFound := db.Get(string(migratorKey))

	if !optFound || o == nil {
		return nil, driver.ErrInvalidMigration
	}

	optVal, ok := o.(option)
	if !ok {
		return nil, driver.ErrInvalidMigration
	}

	return &optVal, nil
}
