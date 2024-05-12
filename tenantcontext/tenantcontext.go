/*
Package tenantcontext provides context keys for the tenant and migration options.
*/
package tenantcontext

import "fmt"

// ContextKey represents a context key which implements the [fmt.Stringer] interface.
type ContextKey interface {
	fmt.Stringer
}

// implement the ContextKey interface.
var _ ContextKey = contextKey{}

// contextKey is the context key for the tenant. It's used as a pointer so it
// fits in an interface{} without allocation; this technique helps avoid
// collisions between packages using context.
type contextKey struct {
	name string
}

// String returns the string representation of the context key.
func (c contextKey) String() string {
	return fmt.Sprintf("gorm-multitenancy/tenantcontext/%s", c.name)
}

var (
	// TenantKey is the context key for the tenant.
	TenantKey = &contextKey{"tenant"}
	// MigrationOptions is the context key for the migration options.
	MigrationOptions = &contextKey{"migration_options"}
)
