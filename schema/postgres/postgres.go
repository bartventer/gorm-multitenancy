package postgres

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/bartventer/gorm-multitenancy/v3/tenantcontext"
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

// SetSearchPath sets the search path for the given database connection to the specified schema.
// It also sets the tenant in the context of the database connection (using the [TenantKey] from the tenantcontext package).
// Additionally, it returns a function that can be used to reset the search path to the default "public" schema and clear the tenant from the context.
// The function takes a *gorm.DB object and a schemaName string as input parameters.
// It returns a modified *gorm.DB object, a ResetSearchPath function, and an error (if any).
// The ResetSearchPath function can be called to clear the tenant from the context and reset the search path to "public".
// If the schemaName is invalid or starts with "pg_", an error will be returned.
//
// Example:
//
//		db, resetSearchPath, err := postgres.SetSearchPath(db, "domain1")
//		if err != nil {
//			fmt.Println(err) // nil
//		}
//		defer resetSearchPath()
//	 // ... do something with the database connection
//
// After calling SetSearchPath, the tenant (in this case, "domain1") will be set in the context of the db object.
// This can be useful for multi-tenant applications where each tenant has its own schema.
//
// [TenantKey]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v3/tenantcontext#TenantKey
func SetSearchPath(db *gorm.DB, schemaName string) (*gorm.DB, ResetSearchPath, error) {
	if !schemaNameRegex.MatchString(schemaName) || pgPrefixRegex.MatchString(schemaName) {
		return nil, nil, fmt.Errorf("invalid schema name")
	}
	db = db.WithContext(context.WithValue(db.Statement.Context, tenantcontext.TenantKey, schemaName)) // set the tenant in the context
	err := db.Exec(fmt.Sprintf("SET search_path TO %s", schemaName)).Error
	if err != nil {
		// return nil, err
		return nil, nil, err
	}
	return db, func() error {
		// clear the tenant from the context
		db = db.WithContext(context.WithValue(db.Statement.Context, tenantcontext.TenantKey, ""))
		// reset the search path to the default "public" schema
		return db.Exec("SET search_path TO public").Error
	}, nil
}
