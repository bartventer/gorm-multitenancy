package irismiddleware

type contextKey struct {
	name string
}

func (c contextKey) String() string {
	return "gorm-multitenancy/middleware/iris/" + c.name
}

var (
	// TenantKey is the key that holds the tenant in a request context.
	TenantKey = &contextKey{"tenant"}
)
