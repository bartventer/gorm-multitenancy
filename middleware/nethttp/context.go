package nethttp

import (
	"fmt"
)

type contextKey struct {
	name string
}

func (c contextKey) String() string {
	return fmt.Sprintf("gorm-multitenancy/middleware/nethttp/%s", c.name)
}

var (
	// TenantKey is the key that holds the tenant in a request context.
	TenantKey = &contextKey{"tenant"}
)
