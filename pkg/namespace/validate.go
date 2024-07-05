// Package namespace provides utilities for validating tenant names in a
// consistent manner across different database systems.
//
// Namespace is a term used to refer to a tenant in the context of the gorm-multitenancy package;
// it is a unique identifier for a tenant and is equivalent to:
//   - PostgreSQL: schema name
//   - MySQL: database name
package namespace

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bartventer/gorm-multitenancy/v8/pkg/gmterrors"
)

// namespaceRegexStr is the regular expression pattern for a valid schema name.
//
// Examples of valid schema names:
//   - "domain1"
//   - "test_domain"
//   - "test123"
//   - "_domain"
const namespaceRegexStr = `^[_a-zA-Z][_a-zA-Z0-9]{2,}$`

var namespaceRegex = regexp.MustCompile(namespaceRegexStr)

const (
	errInvalidPattern = "invalid tenant name: '%s'. Tenant name must match the following pattern: '%s'. This means it must start with an underscore or a letter, followed by at least two characters that can be underscores, letters, or numbers"
	errInvalidPrefix  = "invalid tenant name: %s. Tenant name must not start with 'pg_' as it is reserved for system schemas in PostgreSQL"
)

// checkPattern validates if the tenant name matches the required pattern.
func checkPattern(tenantID string) error {
	if !namespaceRegex.MatchString(tenantID) {
		return fmt.Errorf(errInvalidPattern, tenantID, namespaceRegexStr)
	}
	return nil
}

// checkPrefix validates if the tenant name starts with a reserved prefix.
func checkPrefix(tenantID string) error {
	if strings.HasPrefix(tenantID, "pg_") {
		return fmt.Errorf(errInvalidPrefix, tenantID)
	}
	return nil
}

// Validate validates the tenant name.
func Validate(tenantID string) error {
	if err := checkPattern(tenantID); err != nil {
		return gmterrors.New(err)
	}
	if err := checkPrefix(tenantID); err != nil {
		return gmterrors.New(err)
	}
	return nil
}
