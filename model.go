package multitenancy

// TenantModel a basic GoLang struct which includes the following fields: DomainURL, SchemaName.
// It's intended to be embedded into any public model that needs to be scoped to a tenant, supporting both PostgreSQL and MySQL.
// It may be embedded into your model or you may build your own model without it.
//
// For example:
//
//	type Tenant struct {
//	  multitenancy.TenantModel
//	}
type TenantModel struct {
	// DomainURL is the domain URL of the tenant; same as [net/url.URL.Host].
	DomainURL string `json:"domainURL" mapstructure:"domainURL" gorm:"column:domain_url;uniqueIndex;size:128"`

	// SchemaName is the schema name of the tenant.
	//
	// Field-level permissions are restricted to read and create.
	//
	// The following constraints are applied:
	// 	- unique index
	// 	- size: 63
	//  - check: Not less than 3 characters long
	//
	// Note: Due to differences in regular expression support between PostgreSQL and MySQL,
	// complex validations based on patterns (e.g., ensuring the schema name does not start with 'pg_')
	// should be enforced at the application level or through database-specific features.
	//
	// Examples of valid schema names:
	// 	- "tenant1"
	// 	- "_tenant"
	// 	- "tenant_1"
	//
	// Examples of invalid schema names:
	// 	- "t" (less than 3 characters long)
	// 	- "tenant1!" (contains special characters)
	//
	SchemaName string `json:"schemaName" mapstructure:"schemaName" gorm:"column:schema_name;uniqueIndex;->;<-:create;size:63;check:LENGTH(schema_name) >= 3"`
}

// TenantPKModel is identical to [TenantModel] but with SchemaName as a primary key field.
type TenantPKModel struct {
	// DomainURL is the domain URL of the tenant; same as [net/url.URL.Host].
	DomainURL string `json:"domainURL" mapstructure:"domainURL" gorm:"column:domain_url;uniqueIndex;size:128"`

	// SchemaName is the schema name of the tenant and the primary key of the model.
	// For details on the constraints and rules for this field, see [TenantModel.SchemaName].
	ID string `json:"id" mapstructure:"id" gorm:"column:id;primaryKey;uniqueIndex;->;<-:create;size:63;check:LENGTH(id) >= 3"`
}
