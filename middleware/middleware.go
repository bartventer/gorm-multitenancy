package middleware

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"gorm.io/gorm"
)

const (
	// XTenantHeader is the header key for the tenant
	XTenantHeader = "X-Tenant"
)

// DefaultTenantFromSubdomain extracts the subdomain from the given HTTP request's host.
// It removes the port from the host if present and adds a scheme to the host for parsing.
// The function then parses the URL and extracts the subdomain.
// It returns the extracted subdomain as a string and any error encountered during the process.
//
// This function calls the [ExtractSubdomain] function to extract the subdomain from the host.
//
// [ExtractSubdomain]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v3/middleware#ExtractSubdomain
func DefaultTenantFromSubdomain(r *http.Request) (string, error) {
	// Extract the host from the request
	host := r.Host

	// If the host includes a port, remove it
	if strings.Contains(host, ":") {
		host = strings.Split(host, ":")[0]
	}

	// Add a scheme to the host so it can be parsed by url.Parse
	urlStr := fmt.Sprintf("https://%s", host)

	// Parse the URL
	u, err := url.Parse(urlStr)
	if err != nil {
		return "", err
	}

	// Extract the subdomain
	return ExtractSubdomain(u.String())
}

// DefaultTenantFromHeader extracts the tenant from the X-Tenant header in the HTTP request.
// It returns the extracted tenant as a string and an error if the header is empty or missing.
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
	// ErrTenantInvalid represents an error when the tenant is invalid or not found
	ErrTenantInvalid = fmt.Errorf("invalid tenant or tenant not found")
	// ErrDBInvalid represents an error when the database connection is invalid
	ErrDBInvalid = gorm.ErrInvalidDB
)
