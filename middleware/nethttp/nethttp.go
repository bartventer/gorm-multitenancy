/*
Package nethttp provides a middleware for the [net/http] package, which adds multi-tenancy support.

Example usage:

	import (
	    "net/http"

	    nethttpmw "github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8"
	    "github.com/bartventer/gorm-multitenancy/v8"
	)

	func main() {
	    mux := http.NewServeMux()

	    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	        tenant := r.Context().Value(nethttpmw.TenantKey).(string)
	        fmt.Fprintf(w, "Hello, %s", tenant)
	    })

	    handler := nethttpmw.WithTenant(nethttpmw.DefaultWithTenantConfig)(mux)

	    http.ListenAndServe(":8080", handler)
	}

[net/http]: https://golang.org/pkg/net/http/
*/
package nethttp

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

const (
	// XTenantHeader is the header key for the tenant.
	XTenantHeader = "X-Tenant"
)

var (
	// ErrTenantInvalid represents an error when the tenant is invalid or not found.
	ErrTenantInvalid = errors.New("invalid tenant or tenant not found")
)

// DefaultSkipper represents the default skipper.
func DefaultSkipper(r *http.Request) bool {
	return false
}

// DefaultTenantFromSubdomain extracts the tenant from the subdomain in the HTTP request.
func DefaultTenantFromSubdomain(r *http.Request) (string, error) {
	return ExtractSubdomain(r.Host)
}

// DefaultTenantFromHeader extracts the tenant from the [XTenantHeader] header in the HTTP request.
// It returns the extracted tenant as a string and an error if the header is empty or missing.
func DefaultTenantFromHeader(r *http.Request) (string, error) {
	tenant := r.Header.Get(XTenantHeader)
	tenant = strings.TrimSpace(tenant)
	if tenant == "" {
		return "", fmt.Errorf("%w: failed to get tenant from `%s` header, header is empty", ErrTenantInvalid, XTenantHeader)
	}
	return tenant, nil
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
		ContextKey: TenantKey,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, _ error) {
			http.Error(w, ErrTenantInvalid.Error(), http.StatusInternalServerError)
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
	ContextKey fmt.Stringer

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

	if len(config.TenantGetters) == 0 {
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
