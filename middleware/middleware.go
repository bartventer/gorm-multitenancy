package middleware

import (
	"fmt"
	"net/http"
	"strings"

	multitenancy_ctx "github.com/bartventer/gorm-multitenancy/context"
	"gorm.io/gorm"
)

const (
	// XTenantHeader is the header key for the tenant
	XTenantHeader = "X-Tenant"
)

// DefaultTenantFromSubdomain retrieves the tenant from the request subdomain
func DefaultTenantFromSubdomain(r *http.Request) (string, error) {
	host := r.Host
	host = strings.Replace(host, "127.0.0.1", "", -1) // replace 127.0.0.1, localhost with empty string
	parts := strings.Split(host, ".")
	if len(parts) < 2 {
		return "", fmt.Errorf("failed to get tenant from request host, invalid subdomain")
	}
	tenant := parts[0]

	if tenant == "" {
		return "", fmt.Errorf("failed to get tenant from request host, subdomain is empty")
	}
	return tenant, nil
}

// DefaultTenantFromHeader retrieves the tenant from the request header
func DefaultTenantFromHeader(r *http.Request) (string, error) {
	tenant := r.Header.Get(XTenantHeader)
	tenant = strings.TrimSpace(tenant)
	if tenant == "" {
		return "", fmt.Errorf("failed to get tenant from `%s` header, header is empty", XTenantHeader)
	}
	return tenant, nil
}

// WithTenantConfig represents the config for the tenant middleware
type WithTenantConfig struct {
	DB            *gorm.DB                                // DB is the database connection
	Skipper       func(r *http.Request) bool              // Skipper defines a function to skip middleware
	TenantGetters []func(r *http.Request) (string, error) // TenantGetters gets the tenant from the request; overrides the default getter
}

var (
	// DefaultSkipper represents the default skipper
	DefaultSkipper = func(r *http.Request) bool {
		return false
	}

	// DefaultTenantGetters represents the default tenant getters
	DefaultTenantGetters = []func(r *http.Request) (string, error){
		DefaultTenantFromSubdomain,
		DefaultTenantFromHeader,
	}
)

var (
	// NetHTTPTenantKey is the key that holds the tenant in the request context
	NetHTTPTenantKey = multitenancy_ctx.NetHTTPTenantKey
	// EchoTenantKey is the key that holds the tenant in the echo context
	EchoTenantKey = multitenancy_ctx.EchoTenantKey
)

var (
	// ErrTenantInvalid represents an error when the tenant is invalid or not found
	ErrTenantInvalid = fmt.Errorf("invalid tenant or tenant not found")
	// ErrDBInvalid represents an error when the database connection is invalid
	ErrDBInvalid = gorm.ErrInvalidDB
)
