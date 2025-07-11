/*
Package schema provides utilities for managing PostgreSQL schemas in a multi-tenant application.
*/
package schema

import (
	"errors"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

// SetSearchPath sets the search path for the given database connection to the specified schema name.
// It returns a function that can be used to reset the search path to the default value.
// This function does not perform any validation on the schemaName parameter. It is the
// responsibility of the caller to ensure that the schemaName has been sanitized to avoid SQL
// injection vulnerabilities.
//
// Technically safe for concurrent use by multiple goroutines, but should not be used concurrently
// ito ensuring data integrity and schema isolation. Use a separate database connection or transaction for each
// goroutine that requires a different search path.
//
// Example:
//
//	reset, err := postgres.SetSearchPath(db, "domain1")
//	if err != nil {
//		// handle the error
//	}
//	defer reset() // reset the search path to 'public'
//	// ... do operations with the database with the search path set to 'domain1'
func SetSearchPath(tx *gorm.DB, schemaName string) (reset func() error, err error) {
	tx = tx.Session(&gorm.Session{})
	if schemaName == "" {
		err = errors.New("schema name is empty")
		_ = tx.AddError(err)
		return nil, err
	}
	sql := new(strings.Builder)
	_, _ = sql.WriteString("SET search_path TO ")
	tx.QuoteTo(sql, schemaName)
	if execErr := tx.Exec(sql.String()).Error; execErr != nil {
		err = fmt.Errorf("failed to set search path %q: %w", schemaName, execErr)
		_ = tx.AddError(err)
		return nil, err
	}
	reset = func() error { return tx.Exec("SET search_path TO public").Error }
	return reset, nil
}

// CurrentSearchPath returns the current search path for the given database connection.
func CurrentSearchPath(tx *gorm.DB) string {
	tx = tx.Session(&gorm.Session{})
	var searchPath string
	_ = tx.Raw("SHOW search_path").Scan(&searchPath)
	if searchPath == `"$user", public` {
		return "public"
	}
	return searchPath
}
