package driver

import (
	"cmp"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/bartventer/gorm-multitenancy/v7/pkg/gmterrors"
)

// PublicSchemaEnvVar is the environment variable that contains the name of the public schema.
const PublicSchemaEnvVar = "GMT_PUBLIC_SCHEMA_NAME"

// PublicSchemaName returns the name of the public schema as defined by the [PublicSchemaEnvVar]
// environment variable, defaulting to "public" if the variable is not set. This schema name is
// used to identify shared models.
func PublicSchemaName() string {
	return cmp.Or(os.Getenv(PublicSchemaEnvVar), "public")
}

type (
	// ModelRegistry holds the models registered for multitenancy support, categorizing them into
	// shared and tenant-specific models. Not intended for direct use in application code.
	ModelRegistry struct {
		SharedModels []TenantTabler // SharedModels contains the models that are shared across tenants.
		TenantModels []TenantTabler // TenantModels contains the models that are specific to a tenant.
	}
)

// NewModelRegistry creates and initializes a new ModelRegistry with the provided models, categorizing them into
// shared and tenant-specific based on their characteristics. It returns an error if any model fails validation.
// Not intended for direct use in application code.
func NewModelRegistry(models ...TenantTabler) (*ModelRegistry, error) {
	var (
		registry = &ModelRegistry{
			SharedModels: make([]TenantTabler, 0, len(models)),
			TenantModels: make([]TenantTabler, 0, len(models)),
		}
		errs []error
	)

	for _, model := range models {
		tableName := model.TableName()
		if model.IsSharedModel() {
			if err := validateSharedModel(tableName); err != nil {
				errs = append(errs, err)
				continue
			}
			registry.SharedModels = append(registry.SharedModels, model)
		} else {
			if err := validateTenantModel(tableName); err != nil {
				errs = append(errs, err)
				continue
			}
			registry.TenantModels = append(registry.TenantModels, model)
		}
	}

	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return registry, nil
}

// splitTableName splits a table name into its constituent parts, typically schema and table name.
func splitTableName(tableName string) []string {
	return strings.Split(tableName, ".")
}

// validateSharedModel checks if a shared model's table name conforms to the expected naming convention,
// which includes the public schema prefix. It returns an error if the validation fails.
//
// Example:
//
//	validateSharedModel("public.users") // returns nil
//	validateSharedModel("users") // returns error
func validateSharedModel(tableName string) error {
	parts := splitTableName(tableName)
	public := PublicSchemaName()
	if len(parts) != 2 || parts[0] != public {
		return gmterrors.New(fmt.Errorf("invalid table name for model labeled as public table, table name should start with '%s.', got '%s'", public, tableName))
	}
	return nil
}

// validateTenantModel verifies that a tenant model's table name does not include a schema prefix,
// ensuring it is tenant-specific. It returns an error if the validation fails.
//
// Example:
//
//	validateTenantModel("users") // returns nil
//	validateTenantModel("public.users") // returns error
func validateTenantModel(tableName string) error {
	parts := splitTableName(tableName)
	if len(parts) > 1 {
		return gmterrors.New(fmt.Errorf("invalid table name for model labeled as tenant table, table name should not contain a fullstop, got '%s'", tableName))
	}
	return nil
}
