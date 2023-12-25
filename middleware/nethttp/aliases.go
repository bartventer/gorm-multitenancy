package nethttp

import mw "github.com/bartventer/gorm-multitenancy/middleware"

type (
	// WithTenantConfig defines the config for WithTenant middleware.
	WithTenantConfig = mw.WithTenantConfig
)

var (
	// TenantKey is the key that holds the tenant in a request context.
	TenantKey = mw.NetHTTPTenantKey

	// DefaultSkipper defines the default skipper function.
	DefaultSkipper = mw.DefaultSkipper
	// DefaultTenantGetters represents the default tenant getters
	DefaultTenantGetters = mw.DefaultTenantGetters
)
