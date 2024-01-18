package postgres

import (
	"fmt"
	"strings"

	"gorm.io/gorm"
)

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
	schemaName := getSchemaNameFromSqlExpr(tx.Statement.TableExpr.SQL)
	// if the schema name is empty, return an error
	if schemaName == "" {
		return "", fmt.Errorf("schema name is empty")
	}
	return schemaName, nil
}

// getSchemaNameFromSqlExpr extracts the schema name from a SQL expression.
// It splits the input string by the dot; if the length is 1, then there is no schema name.
// Otherwise, it retrieves the first element and removes any backslashes and double quotes before returning the schema name.
//
// Example:
//
//	"\"test_domain\".\"mock_private\"" -> "test_domain"
//	"\"mock_private\"" -> ""
func getSchemaNameFromSqlExpr(tableExprSql string) string {
	// split the string by the dot
	split := strings.Split(tableExprSql, ".")
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
