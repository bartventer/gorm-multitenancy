/*
Package irismiddleware provides a middleware for the Iris framework, which adds multi-tenancy support.

Example usage:

	import (

		"github.com/kataras/iris/v12"
		irismiddleware "github.com/bartventer/gorm-multitenancy/middleware/iris/v8"

	)

	func main() {
	    app := iris.New()

	    app.Use(irismiddleware.WithTenant(irismiddleware.DefaultWithTenantConfig))

	    app.Get("/", func(ctx iris.Context) {
	        tenant := ctx.Values().GetString(irismiddleware.TenantKey.String())
	        ctx.WriteString("Hello, " + tenant)
	    })

	    app.Listen(":8080")
	}
*/
package irismiddleware

import (
	"fmt"
	"net/http"

	nethttpmw "github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8"
	"github.com/kataras/iris/v12"
)

// DefaultSkipper returns false which processes the middleware.
func DefaultSkipper(ctx iris.Context) bool {
	return nethttpmw.DefaultSkipper(ctx.Request())
}

// DefaultTenantFromSubdomain extracts the subdomain from the given HTTP request's host.
func DefaultTenantFromSubdomain(ctx iris.Context) (string, error) {
	return nethttpmw.DefaultTenantFromSubdomain(ctx.Request())
}

// DefaultTenantFromHeader extracts the tenant from the header in the HTTP request.
func DefaultTenantFromHeader(ctx iris.Context) (string, error) {
	return nethttpmw.DefaultTenantFromHeader(ctx.Request())
}

var (
	// DefaultWithTenantConfig is the default configuration for the WithTenant middleware.
	DefaultWithTenantConfig = WithTenantConfig{
		Skipper: DefaultSkipper,
		TenantGetters: []func(ctx iris.Context) (string, error){
			DefaultTenantFromSubdomain,
			DefaultTenantFromHeader,
		},
		ContextKey: TenantKey,
		ErrorHandler: func(ctx iris.Context, _ error) {
			ctx.StopWithJSON(http.StatusInternalServerError, iris.Map{"error": nethttpmw.ErrTenantInvalid.Error()})
		},
	}
)

// WithTenantConfig represents the configuration options for the tenant middleware in Iris.
type WithTenantConfig struct {
	// Skipper defines a function to skip the middleware.
	Skipper func(ctx iris.Context) bool

	// TenantGetters is a list of functions that retrieve the tenant from the request.
	// Each function should return the tenant as a string and an error if any.
	// The functions are executed in order until a valid tenant is found.
	TenantGetters []func(ctx iris.Context) (string, error)

	// ContextKey is the key used to store the tenant in the context.
	ContextKey fmt.Stringer

	// ErrorHandler is a callback function that is called when an error occurs during the tenant retrieval process.
	ErrorHandler func(ctx iris.Context, err error)

	// SuccessHandler is a callback function that is called after the tenant is successfully set in the Iris context.
	// It can be used to perform additional operations, such as modifying the database connection based on the tenant.
	SuccessHandler func(ctx iris.Context)
}

// WithTenant returns a new tenant middleware with the provided configuration.
// If the configuration is not provided, the [DefaultWithTenantConfig] is used.
func WithTenant(config WithTenantConfig) iris.Handler {
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

	return func(ctx iris.Context) {
		if config.Skipper(ctx) {
			ctx.Next()
			return
		}

		var tenant string
		var err error
		for _, getter := range config.TenantGetters {
			tenant, err = getter(ctx)
			if err == nil {
				break
			}
		}
		if err != nil {
			config.ErrorHandler(ctx, err)
			return
		}

		ctx.Values().Set(config.ContextKey.String(), tenant)

		if config.SuccessHandler != nil {
			config.SuccessHandler(ctx)
		}

		ctx.Next()
	}
}
