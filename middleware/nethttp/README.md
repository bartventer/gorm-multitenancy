# NetHTTP Middleware for Multitenancy

NetHTTP middleware provides multitenancy support for the [net/http](https://golang.org/pkg/net/http/) package in Go.

For valid tenant information, it calls the next handler. For missing or invalid tenant information, it sends "500 - Internal Server Error" response with the error message "Invalid tenant or tenant not found".

## Usage

```go
import (
    "net/http"

    nethttpmw "github.com/bartventer/gorm-multitenancy/v6/middleware/nethttp"
    "github.com/bartventer/gorm-multitenancy/v6/tenantcontext"
)

func main() {
    mux := http.NewServeMux()

    mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        tenant := r.Context().Value(tenantcontext.TenantKey).(string)
        fmt.Fprintf(w, "Hello, %s", tenant)
    })

    handler := nethttpmw.WithTenant(nethttpmw.DefaultWithTenantConfig)(mux)

    http.ListenAndServe(":8080", handler)
}
```

## Configuration

```go
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
```

### Default Configuration

```go
var DefaultWithTenantConfig = WithTenantConfig{
		Skipper: DefaultSkipper,
		TenantGetters: []func(r *http.Request) (string, error){
			DefaultTenantFromSubdomain,
			DefaultTenantFromHeader,
		},
		ContextKey: tenantcontext.TenantKey,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, _ error) {
			http.Error(w, ErrTenantInvalid.Error(), http.StatusInternalServerError)
		},
	}
```