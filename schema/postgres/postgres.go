package postgres

import (
	"fmt"
	"regexp"
	"strings"

	"gorm.io/gorm"
)

// ResetSearchPath is a function that resets the search path to the default value.
type ResetSearchPath func() error

const (
	schemaNameRegexStr = `^[_a-zA-Z][_a-zA-Z0-9]{2,}$`
	pgPrefixRegexStr   = `^pg_`
)

var (
	schemaNameRegex = regexp.MustCompile(schemaNameRegexStr)
	pgPrefixRegex   = regexp.MustCompile(pgPrefixRegexStr)
)

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
	// to avoid
	if !schemaNameRegex.MatchString(schemaName) || pgPrefixRegex.MatchString(schemaName) {
		_ = db.AddError(fmt.Errorf("invalid schema name; schema name must match the regex %s and must not start with 'pg_'", schemaNameRegexStr))
		return db, reset
	}
	if err := db.Exec(fmt.Sprintf("SET search_path TO %s", db.Statement.Quote(schemaName))).Error; err != nil {
		_ = db.AddError(err)
		return db, reset
	}
	reset = func() error {
		return db.Exec("SET search_path TO public").Error
	}
	return db, reset
}

// GetSchemaNameFromDb retrieves the schema name from the given gorm.DB transaction.
// It first checks if the table expression is not nil, then extracts the schema name from the table expression SQL.
// If the schema name is empty, it returns an error.
//
// It is intended to be used in a gorm hook, such as BeforeCreate, BeforeUpdate, etc.
//
// Example:
//
//	type User struct {
//		gorm.Model
//		Username string
//	}
//
//	func (User) TableName() string {
//		return "domain1.mock_private"
//	}
//
//	func (User) BeforeCreate(tx *gorm.DB) (err error) {
//		schemaName, err := postgres.GetSchemaNameFromDb(tx) // schemaName = "domain1"
//		if err != nil {
//			return err
//		}
//		// ... do something with schemaName
//		return nil
//	}
func GetSchemaNameFromDb(tx *gorm.DB) (string, error) {
	// get the table expression sql
	if tx.Statement.TableExpr == nil {
		return "", fmt.Errorf("table expression is nil")
	}
	// get the schema name from the table expression sql
	schemaName := getSchemaNameFromSQLExpr(tx.Statement.TableExpr.SQL)
	// if the schema name is empty, return an error
	if schemaName == "" {
		return "", fmt.Errorf("schema name is empty")
	}
	return schemaName, nil
}

// getSchemaNameFromSQLExpr extracts the schema name from a SQL expression.
// It splits the input string by the dot; if the length is 1, then there is no schema name.
// Otherwise, it retrieves the first element and removes any backslashes and double quotes before returning the schema name.
//
// Example:
//
//	"\"test_domain\".\"mock_private\"" -> "test_domain"
//	"\"mock_private\"" -> ""
func getSchemaNameFromSQLExpr(tableExprSQL string) string {
	// split the string by the dot
	split := strings.Split(tableExprSQL, ".")
	// if the length is 1, then there is no schema name
	if len(split) == 1 {
		return ""
	}
	// get the first element
	schemaName := split[0]
	// remove the backslash and double quotes
	schemaName = strings.ReplaceAll(schemaName, "\"", "")
	return schemaName
}
