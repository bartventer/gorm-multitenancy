package echo

import multitenancy "github.com/bartventer/gorm-multitenancy/v7"

var (
	// TenantKey is the key that holds the tenant in a request context.
	TenantKey = multitenancy.TenantKey
)
