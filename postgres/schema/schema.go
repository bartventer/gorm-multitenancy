/*
Package schema provides utilities for managing PostgreSQL schemas in a multi-tenant application.
*/
package schema

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// SetSearchPath sets the search path for the given database connection to the specified schema name.
// It returns a function that can be used to reset the search path to the default value.
// This function does not perform any validation on the schemaName parameter. It is the
// responsibility of the caller to ensure that the schemaName has been sanitized to avoid SQL
// injection vulnerabilities.
//
// Not safe for concurrent use by multiple goroutines. Use a separate database connection or
// transaction for each goroutine that requires a different search path.
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
		tx.AddError(err)
		return nil, err
	}
	sqlstr := "SET search_path TO " + schemaName
	if execErr := tx.Exec(sqlstr).Error; execErr != nil {
		err = fmt.Errorf("failed to set search path %q: %w", schemaName, execErr)
		tx.AddError(err)
		return nil, err
	}
	reset = func() error {
		return tx.Exec("SET search_path TO public").Error
	}
	return reset, nil
}

// CurrentSearchPath returns the current search path for the given database connection.
func CurrentSearchPath(tx *gorm.DB) string {
	tx = tx.Session(&gorm.Session{})
	var searchPath string
	tx.Raw("SHOW search_path").Scan(&searchPath)
	if searchPath == `"$user", public` {
		return "public"
	}
	return searchPath
}
