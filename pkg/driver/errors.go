package driver

import (
	"errors"
)

var (
	// ErrInvalidMigration is returned when an invalid migration is detected.
	ErrInvalidMigration = errors.New(`
		Invalid migration. Please ensure you're using MigrateSharedModels or MigrateTenantModels
		instead of calling AutoMigrate directly.
	`)
)
