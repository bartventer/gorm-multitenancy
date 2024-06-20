# Echo Middleware for Multitenancy

[![Go Reference](https://pkg.go.dev/badge/github.com/bartventer/gorm-multitenancy.svg)](https://pkg.go.dev/github.com/bartventer/gorm-multitenancy/middleware/echo/v7)

Echo middleware provides multitenancy support for the [Echo](https://echo.labstack.com/docs) framework.

For valid tenant information, it calls the next handler.
For missing or invalid tenant information, it sends "500 - Internal Server Error" response with the error message "Invalid tenant".

## Usage

```go
import (
    "net/http"

    echomw "github.com/bartventer/gorm-multitenancy/middleware/echo/v7"
    "github.com/bartventer/gorm-multitenancy/v7"
    "github.com/labstack/echo/v4"
)

func main() {
    e := echo.New()

    e.Use(echomw.WithTenant(echomw.DefaultWithTenantConfig))

    e.GET("/", func(c echo.Context) error {
        tenant := c.Get(multitenancy.TenantKey).(string)
        return c.String(http.StatusOK, "Hello, "+tenant)
    })

    e.Start(":8080")
}
```

## Configuration

```go
type WithTenantConfig struct {

	// Skipper defines a function to skip the middleware.
	Skipper func(c echo.Context) bool

	// TenantGetters is a list of functions that retrieve the tenant from the request.
	// Each function should return the tenant as a string and an error if any.
	// The functions are executed in order until a valid tenant is found.
	TenantGetters []func(c echo.Context) (string, error)

	// ContextKey is the key used to store the tenant in the context.
	ContextKey multitenancy.ContextKey

	// ErrorHandler is a callback function that is called when an error occurs during the tenant retrieval process.
	ErrorHandler func(c echo.Context, err error) error

	// SuccessHandler is a callback function that is called after the tenant is successfully set in the echo context.
	// It can be used to perform additional operations, such as modifying the database connection based on the tenant.
	SuccessHandler func(c echo.Context)
}
```

### Default Configuration

```go
var	DefaultWithTenantConfig = WithTenantConfig{
		Skipper: DefaultSkipper,
		TenantGetters: []func(c echo.Context) (string, error){
			DefaultTenantFromSubdomain,
			DefaultTenantFromHeader,
		},
		ContextKey: multitenancy.TenantKey,
		ErrorHandler: func(c echo.Context, _ error) error {
			return echo.NewHTTPError(http.StatusInternalServerError, nethttpmw.ErrTenantInvalid.Error())
		},
	}
```