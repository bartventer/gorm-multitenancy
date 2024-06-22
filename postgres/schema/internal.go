package schema

import "strings"

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
