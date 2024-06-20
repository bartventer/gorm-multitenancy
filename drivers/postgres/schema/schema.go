/*
Package schema provides utilities for managing PostgreSQL schemas in a multi-tenant application.
*/
package schema

import (
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// schemaNameRegexStr is the regular expression pattern for a valid schema name.
//
// Examples of valid schema names:
//   - "domain1"
//   - "test_domain"
//   - "test123"
//   - "_domain"
const schemaNameRegexStr = `^[_a-zA-Z][_a-zA-Z0-9]{2,}$`

var schemaNameRegex = regexp.MustCompile(schemaNameRegexStr)

// ValidateTenantName checks the validity of a provided tenant name.
// A tenant name is considered valid if it:
//  1. Matches the pattern `^[_a-zA-Z][_a-zA-Z0-9]{2,}$`. This means it must start with an underscore or a letter, followed by at least two characters that can be underscores, letters, or numbers.
//  2. Does not start with "pg_". The prefix "pg_" is reserved for system schemas in PostgreSQL.
//
// If the tenant name is invalid, the function returns an error with a detailed explanation.
func ValidateTenantName(tenant string) error {
	if !schemaNameRegex.MatchString(tenant) {
		return fmt.Errorf(`
invalid tenant name: '%s'. Tenant name must match the following pattern: '%s'.
This means it must start with an underscore or a letter, followed by at least two characters that can be underscores, letters, or numbers`,
			tenant, schemaNameRegexStr)
	}
	if strings.HasPrefix(tenant, "pg_") {
		return fmt.Errorf("invalid tenant name: %s. Tenant name must not start with 'pg_' as it is reserved for system schemas in PostgreSQL", tenant)
	}
	return nil
}

// ResetSearchPath is a function that resets the search path to the default value.
type ResetSearchPath func() error

// SetSearchPath sets the search path for the given database connection to the specified schema name.
// It returns the modified database connection and a function that can be used to reset the search path to the default 'public' schema.
// If the schema name is invalid or starts with 'pg_', an error is added to the database connection's error list.
//
// Example:
//
//	db, reset := postgres.SetSearchPath(db, "domain1")
//	if db.Error != nil {
//		// handle the error
//	}
//	defer reset() // reset the search path to 'public'
//	// ... do something with the database connection (with the search path set to 'domain1')
func SetSearchPath(db *gorm.DB, schemaName string) (*gorm.DB, ResetSearchPath) {
	var reset ResetSearchPath
	if err := ValidateTenantName(schemaName); err != nil {
		_ = db.AddError(err)
		return db, reset
	}
	sql := "SET search_path TO " + db.Statement.Quote(schemaName)
	if err := db.Exec(sql).Error; err != nil {
		_ = db.AddError(err)
		return db, reset
	}
	reset = func() error {
		return db.Exec("SET search_path TO public").Error
	}
	return db, reset
}
