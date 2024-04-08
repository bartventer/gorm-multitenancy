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

	// SchemaName is the schema name of the tenant.
	//
	// Field-level permissions are restricted to read and create.
	//
	// The following constraints are applied:
	// 	- unique index
	// 	- size: 63
	// 	- check: schema_name ~ '^[_a-zA-Z][_a-zA-Z0-9]{2,}$' AND schema_name !~ '^pg_' (to prevent invalid schema names)
	SchemaName string `json:"schemaName" gorm:"column:schema_name;uniqueIndex;->;<-:create;size:63;check:schema_name ~ '^[_a-zA-Z][_a-zA-Z0-9]{2,}$' AND schema_name !~ '^pg_'"`
}
