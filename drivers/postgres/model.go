package postgres

// TenantModel a basic GoLang struct which includes the following fields: DomainURL, SchemaName.
// It's intended to be embedded into any public postgresql model that needs to be scoped to a tenant.
// It may be embedded into your model or you may build your own model without it.
//
// For example:
//
//	type Tenant struct {
//	  postgres.TenantModel
//	}
type TenantModel struct {
	// DomainURL is the domain URL of the tenant
	DomainURL string `json:"domainURL" gorm:"column:domain_url;uniqueIndex;size:128"`

	// SchemaName is the schema name of the tenant and the primary key of the model.
	//
	// Field-level permissions are restricted to read and create.
	//
	// The following constraints are applied:
	// 	- unique index
	// 	- size: 63
	//
	// Additionally, check constraints are applied to ensure that the schema name adheres to the following rules:
	// 	- It must start with an underscore or a letter.
	// 	- The rest of the string can contain underscores, letters, and numbers.
	// 	- It must be at least 3 characters long.
	// 	- It must not start with 'pg_', as this prefix is reserved for system schemas.
	//
	// Examples of valid schema names:
	// 	- "tenant1"
	// 	- "_tenant"
	// 	- "tenant_1"
	//
	// Examples of invalid schema names:
	// 	- "1tenant" (does not start with an underscore or a letter)
	// 	- "pg_tenant" (starts with 'pg_')
	// 	- "t" (less than 3 characters long)
	//
	SchemaName string `json:"schemaName" gorm:"column:schema_name;uniqueIndex;->;<-:create;size:63;check:schema_name ~ '^[_a-zA-Z][_a-zA-Z0-9]{2,}$' AND schema_name !~ '^pg_'"`
}

// TenantPKModel is identical to [TenantModel] but with SchemaName as a primary key field.
type TenantPKModel struct {
	// DomainURL is the domain URL of the tenant
	DomainURL string `json:"domainURL" gorm:"column:domain_url;uniqueIndex;size:128"`

	// SchemaName is the schema name of the tenant and the primary key of the model.
	// For details on the constraints and rules for this field, see [TenantModel.SchemaName].
	SchemaName string `json:"schemaName" gorm:"column:schema_name;primaryKey;uniqueIndex;->;<-:create;size:63;check:schema_name ~ '^[_a-zA-Z][_a-zA-Z0-9]{2,}$' AND schema_name !~ '^pg_'"`
}
