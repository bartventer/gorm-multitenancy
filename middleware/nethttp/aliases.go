package nethttp

import "github.com/bartventer/gorm-multitenancy/v5/tenantcontext"

var (
	// TenantKey is the key that holds the tenant in a request context.
	TenantKey = tenantcontext.TenantKey
)
