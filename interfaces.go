package multitenancy

import (
	"gorm.io/gorm"
)

// TenantTabler is the interface for tenant tables.
type TenantTabler interface {
	// IsTenantTable returns true if the table is a tenant table
	IsTenantTable() bool
}

// Migrator is an alias for [gorm.Migrator].
//
// [gorm.Migrator]: https://pkg.go.dev/gorm.io/gorm#Migrator
type Migrator = gorm.Migrator
