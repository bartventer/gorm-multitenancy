package driver

import (
	"errors"
)

var (
	// ErrInvalidMigration is returned when an invalid migration is detected.
	ErrInvalidMigration = errors.Join(
		errors.New("invalid migration"),
		errors.New("please ensure you're using MigrateSharedModels or MigrateTenantModels instead of calling AutoMigrate directly"),
	)
)
