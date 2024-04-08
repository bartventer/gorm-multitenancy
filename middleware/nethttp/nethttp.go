// Package nethttp provides a middleware for the net/http package.
package nethttp

import (
	"context"
	"net/http"

	mw "github.com/bartventer/gorm-multitenancy/v5/middleware"
	"github.com/bartventer/gorm-multitenancy/v5/tenantcontext"
)

// DefaultSkipper returns false which processes the middleware.
// It calls the default [DefaultSkipper] function to determine if the middleware should be skipped.
func DefaultSkipper(r *http.Request) bool {
	return mw.DefaultSkipper(r)
}

// DefaultTenantFromSubdomain extracts the subdomain from the given HTTP request's
// host. It calls the default [DefaultTenantFromSubdomain] function to extract the subdomain from the host.
func DefaultTenantFromSubdomain(r *http.Request) (string, error) {
	return mw.DefaultTenantFromSubdomain(r)
}

// DefaultTenantFromHeader extracts the tenant from the X-Tenant header in the HTTP request.
// It calls the default [DefaultTenantFromHeader] function to extract the tenant from the header.
func DefaultTenantFromHeader(r *http.Request) (string, error) {
	return mw.DefaultTenantFromHeader(r)
}

var (
	// DefaultWithTenantConfig is the default configuration for the WithTenant middleware.
	// It uses the default skipper, tenant getters, context key, and error handler.
	DefaultWithTenantConfig = WithTenantConfig{
		Skipper: DefaultSkipper,
		TenantGetters: []func(r *http.Request) (string, error){
			DefaultTenantFromSubdomain,
			DefaultTenantFromHeader,
		},
		ContextKey: tenantcontext.TenantKey,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, _ error) {
			http.Error(w, mw.ErrTenantInvalid.Error(), http.StatusInternalServerError)
		},
	}
)

// WithTenantConfig represents the configuration options for the tenant middleware in net/http.
type WithTenantConfig struct {
	// Skipper defines a function to skip the middleware.
	Skipper func(r *http.Request) bool

	// TenantGetters is a list of functions that retrieve the tenant from the request.
	// Each function should return the tenant as a string and an error if any.
	// The functions are executed in order until a valid tenant is found.
	TenantGetters []func(r *http.Request) (string, error)

	// ContextKey is the key used to store the tenant in the context.
	ContextKey tenantcontext.ContextKey

	// ErrorHandler is a callback function that is called when an error occurs during the tenant retrieval process.
	ErrorHandler func(w http.ResponseWriter, r *http.Request, err error)

	// SuccessHandler is a callback function that is called after the tenant is successfully set in the http context.
	// It can be used to perform additional operations, such as modifying the database connection based on the tenant.
	SuccessHandler func(w http.ResponseWriter, r *http.Request)
}

// WithTenant is a middleware function that adds multi-tenancy support to a net/http application.
// It takes a WithTenantConfig struct as input and returns a http.Handler.
// The WithTenantConfig struct allows customization of the middleware behavior.
// The middleware checks if the request should be skipped based on the Skipper function.
// It retrieves the tenant information using the TenantGetters functions.
// If an error occurs while retrieving the tenant, the ErrorHandler function is called.
// The retrieved tenant is then set in the request context using the ContextKey.
// Finally, the SuccessHandler function is called if provided, and the next handler is invoked.
func WithTenant(config WithTenantConfig) func(http.Handler) http.Handler {
	if config.Skipper == nil {
		config.Skipper = DefaultWithTenantConfig.Skipper
	}

	if config.TenantGetters == nil {
		config.TenantGetters = DefaultWithTenantConfig.TenantGetters
	}

	if config.ContextKey == nil {
		config.ContextKey = DefaultWithTenantConfig.ContextKey
	}

	if config.ErrorHandler == nil {
		config.ErrorHandler = DefaultWithTenantConfig.ErrorHandler
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if config.Skipper(r) {
				next.ServeHTTP(w, r)
				return
			}
			var (
				tenant string
				err    error
			)
			for _, getter := range config.TenantGetters {
				tenant, err = getter(r)
				if err == nil {
					break
				}
			}
			if err != nil {
				config.ErrorHandler(w, r, err)
				return
			}
			// set tenant in request context
			ctx := context.WithValue(r.Context(), config.ContextKey, tenant)
			r = r.WithContext(ctx)

			// call success handler
			if config.SuccessHandler != nil {
				config.SuccessHandler(w, r)
			}
			next.ServeHTTP(w, r)
		})
	}
}
