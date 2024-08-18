/*
Package ginmiddleware provides a middleware for the Gin framework, which adds multi-tenancy support.

Example usage:

	import (
	    "net/http"

	    ginmiddleware "github.com/bartventer/gorm-multitenancy/middleware/gin/v8"
	    "github.com/bartventer/gorm-multitenancy/v8"
	    "github.com/gin-gonic/gin"
	)

	func main() {
	    r := gin.Default()

	    r.Use(ginmiddleware.WithTenant(ginmiddleware.DefaultWithTenantConfig))

	    r.GET("/", func(c *gin.Context) {
	        tenant := c.GetString(ginmiddleware.TenantKey)
	        c.String(http.StatusOK, "Hello, "+tenant)
	    })

	    r.Run(":8080")
	}
*/
package ginmiddleware

import (
	"fmt"
	"net/http"

	nethttpmw "github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8"
	"github.com/gin-gonic/gin"
)

// DefaultSkipper returns false which processes the middleware.
func DefaultSkipper(c *gin.Context) bool {
	return nethttpmw.DefaultSkipper(c.Request)
}

// DefaultTenantFromSubdomain extracts the subdomain from the given HTTP request's host.
func DefaultTenantFromSubdomain(c *gin.Context) (string, error) {
	return nethttpmw.DefaultTenantFromSubdomain(c.Request)
}

// DefaultTenantFromHeader extracts the tenant from the header in the HTTP request.
func DefaultTenantFromHeader(c *gin.Context) (string, error) {
	return nethttpmw.DefaultTenantFromHeader(c.Request)
}

var (
	// DefaultWithTenantConfig is the default configuration for the WithTenant middleware.
	DefaultWithTenantConfig = WithTenantConfig{
		Skipper: DefaultSkipper,
		TenantGetters: []func(c *gin.Context) (string, error){
			DefaultTenantFromSubdomain,
			DefaultTenantFromHeader,
		},
		ContextKey: TenantKey,
		ErrorHandler: func(c *gin.Context, _ error) {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": nethttpmw.ErrTenantInvalid.Error()})
		},
	}
)

// WithTenantConfig represents the configuration options for the tenant middleware in Gin.
type WithTenantConfig struct {

	// Skipper defines a function to skip the middleware.
	Skipper func(c *gin.Context) bool

	// TenantGetters is a list of functions that retrieve the tenant from the request.
	// Each function should return the tenant as a string and an error if any.
	// The functions are executed in order until a valid tenant is found.
	TenantGetters []func(c *gin.Context) (string, error)

	// ContextKey is the key used to store the tenant in the context.
	ContextKey fmt.Stringer

	// ErrorHandler is a callback function that is called when an error occurs during the tenant retrieval process.
	ErrorHandler func(c *gin.Context, err error)

	// SuccessHandler is a callback function that is called after the tenant is successfully set in the Gin context.
	// It can be used to perform additional operations, such as modifying the database connection based on the tenant.
	SuccessHandler func(c *gin.Context)
}

// WithTenant is a middleware function that adds multi-tenancy support to a Gin application.
func WithTenant(config WithTenantConfig) gin.HandlerFunc {
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

	return func(c *gin.Context) {
		if config.Skipper(c) {
			c.Next()
			return
		}

		var tenant string
		var err error
		for _, getter := range config.TenantGetters {
			tenant, err = getter(c)
			if err == nil {
				break
			}
		}
		if err != nil {
			config.ErrorHandler(c, err)
			return
		}

		c.Set(config.ContextKey.String(), tenant)

		if config.SuccessHandler != nil {
			config.SuccessHandler(c)
		}

		c.Next()
	}
}
