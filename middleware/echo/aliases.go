package echo

import (
	"github.com/bartventer/gorm-multitenancy/v3/tenantcontext"
)

var (
	// TenantKey is the key that holds the tenant in a request context.
	TenantKey = tenantcontext.TenantKey
)
