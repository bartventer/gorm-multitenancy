package ginmiddleware

import "github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8"

// XTenantHeader is an alias for [nethttp.XTenantHeader].
const XTenantHeader = nethttp.XTenantHeader

// ErrTenantInvalid is an alias for [nethttp.ErrTenantInvalid].
var ErrTenantInvalid = nethttp.ErrTenantInvalid
