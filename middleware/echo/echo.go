// Package echo provides a middleware for the Echo framework.
package echo

import (
	"net/http"

	mw "github.com/bartventer/gorm-multitenancy/v4/middleware"
	"github.com/bartventer/gorm-multitenancy/v4/tenantcontext"
	"github.com/labstack/echo/v4"
)

// DefaultSkipper returns false which processes the middleware.
// It calls the default [DefaultSkipper] function to determine if the middleware should be skipped.
//
// [DefaultSkipper]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v4/middleware#DefaultSkipper
func DefaultSkipper(c echo.Context) bool {
	return mw.DefaultSkipper(c.Request())
}

// DefaultTenantFromSubdomain extracts the subdomain from the given HTTP request's
// host. It calls the default [DefaultTenantFromSubdomain] function to extract the subdomain from the host.
//
// [DefaultTenantFromSubdomain]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v4/middleware#DefaultTenantFromSubdomain
func DefaultTenantFromSubdomain(c echo.Context) (string, error) {
	return mw.DefaultTenantFromSubdomain(c.Request())
}

// DefaultTenantFromHeader extracts the tenant from the X-Tenant header in the HTTP request.
// It calls the default [DefaultTenantFromHeader] function to extract the tenant from the header.
//
// [DefaultTenantFromHeader]: https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/v4/middleware#DefaultTenantFromHeader
func DefaultTenantFromHeader(c echo.Context) (string, error) {
	return mw.DefaultTenantFromHeader(c.Request())
}

var (
	// DefaultWithTenantConfig is the default configuration for the WithTenant middleware.
	// It uses the default skipper, tenant getters, context key, and error handler.
	DefaultWithTenantConfig = WithTenantConfig{
		Skipper: DefaultSkipper,
		TenantGetters: []func(c echo.Context) (string, error){
			DefaultTenantFromSubdomain,
			DefaultTenantFromHeader,
		},
		ContextKey: tenantcontext.TenantKey,
		ErrorHandler: func(c echo.Context, _ error) error {
			return echo.NewHTTPError(http.StatusInternalServerError, mw.ErrTenantInvalid.Error())
		},
	}
)

// WithTenantConfig represents the configuration options for the tenant middleware in Echo.
type WithTenantConfig struct {

	// Skipper defines a function to skip the middleware.
	Skipper func(c echo.Context) bool

	// TenantGetters is a list of functions that retrieve the tenant from the request.
	// Each function should return the tenant as a string and an error if any.
	// The functions are executed in order until a valid tenant is found.
	TenantGetters []func(c echo.Context) (string, error)

	// ContextKey is the key used to store the tenant in the context.
	ContextKey tenantcontext.ContextKey

	// ErrorHandler is a callback function that is called when an error occurs during the tenant retrieval process.
	ErrorHandler func(c echo.Context, err error) error

	// SuccessHandler is a callback function that is called after the tenant is successfully set in the echo context.
	// It can be used to perform additional operations, such as modifying the database connection based on the tenant.
	SuccessHandler func(c echo.Context)
}

// WithTenant is a middleware function that adds multi-tenancy support to an Echo application.
// It takes a WithTenantConfig struct as input and returns an Echo MiddlewareFunc.
// The WithTenantConfig struct allows customization of the middleware behavior.
// The middleware checks if the request should be skipped based on the Skipper function.
// It retrieves the tenant information using the TenantGetters functions.
// If an error occurs while retrieving the tenant, the ErrorHandler function is called.
// The retrieved tenant is then set in the request context using the ContextKey.
// Finally, the SuccessHandler function is called if provided, and the next handler is invoked.
func WithTenant(config WithTenantConfig) echo.MiddlewareFunc {

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

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c) {
				return next(c)
			}
			var (
				tenant string
				err    error
			)
			for _, getter := range config.TenantGetters {
				tenant, err = getter(c)
				if err == nil {
					break
				}
			}
			if err != nil {
				return config.ErrorHandler(c, err)
			}
			// set tenant in request context
			c.Set(config.ContextKey.String(), tenant)

			// call success handler
			if config.SuccessHandler != nil {
				config.SuccessHandler(c)
			}
			return next(c)
		}
	}
}
