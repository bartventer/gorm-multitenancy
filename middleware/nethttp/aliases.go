package nethttp

import multitenancy "github.com/bartventer/gorm-multitenancy/v6"

var (
	// TenantKey is the key that holds the tenant in a request context.
	TenantKey = multitenancy.TenantKey
)
