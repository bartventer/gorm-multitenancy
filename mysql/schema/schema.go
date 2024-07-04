/*
Package schema provides utilities for managing MySQL databases in a multi-tenant application.

The term schema is used interchangeably with database in MySQL for the purposes of this package.
*/
package schema

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// UseDatabase sets the database for the given connection to the specified database name.
// It returns a function that can be used to reset the database to the default value.
// This function does not perform any validation on the dbName parameter. It is the
// responsibility of the caller to ensure that the dbName has been sanitized to avoid SQL
// injection vulnerabilities.
//
// Example:
//
//	reset, err := schema.UseDatabase(db, "domain1")
//	if err != nil {
//		// handle the error
//	}
//	defer reset() // reset the database to 'public'
//	// ... do operations with the database with the database set to 'domain1'
func UseDatabase(tx *gorm.DB, dbName string) (func() error, error) {
	tx = tx.Session(&gorm.Session{})
	if dbName == "" {
		return nil, errors.New("database name is empty")
	}

	if err := tx.Exec("USE " + tx.Statement.Quote(dbName)).Error; err != nil {
		return nil, fmt.Errorf("failed to set database: %w", err)
	}

	return func() error {
		return tx.Exec("USE public").Error
	}, nil
}
