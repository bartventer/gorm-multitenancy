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
// Not safe for concurrent use by multiple goroutines. Use a separate database connection or
// transaction for each goroutine that requires a different database.
//
// Example:
//
//	reset, err := schema.UseDatabase(db, "domain1")
//	if err != nil {
//		// handle the error
//	}
//	defer reset() // reset the database to 'public'
//	// ... do operations with the database with the database set to 'domain1'
func UseDatabase(tx *gorm.DB, dbName string) (reset func() error, err error) {
	if dbName == "" {
		err = errors.New("database name is empty")
		tx.AddError(err)
		return nil, err
	}

	sqlstr := "USE " + dbName
	if execErr := tx.Exec(sqlstr).Error; execErr != nil {
		err = fmt.Errorf("failed to set database %q: %w", dbName, execErr)
		tx.AddError(err)
		return nil, err
	}

	reset = func() error {
		return tx.Exec("USE public").Error
	}
	return reset, nil
}
