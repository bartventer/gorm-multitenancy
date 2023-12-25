package context

import "fmt"

// contextKey is the context key for the tenant. It's used as a pointer so it
// fits in an interface{} without allocation; this technique helps avoid
// collisions between packages using context.
type contextKey struct {
	name string
}

// String returns the string representation of the context key.
func (c contextKey) String() string {
	return fmt.Sprintf("multitenancy/%s", c.name)
}

var (
	// NetHTTPTenantKey is the context key for the tenant (net/http)
	NetHTTPTenantKey = &contextKey{"nethttp/tenant"}
	// EchoTenantKey is the context key for the tenant (echo)
	EchoTenantKey = &contextKey{"echo/tenant"}
	// DBTenantKey is the context key for the tenant (gorm.DB)
	DBTenantKey = &contextKey{"db/tenant"}
	// MultitenantMigrationOptions is the context key for the migration options
	MultitenantMigrationOptions = &contextKey{"db/options"}
)
