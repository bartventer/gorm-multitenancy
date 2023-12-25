package nethttp

import (
	"context"
	"net/http"

	mw "github.com/bartventer/gorm-multitenancy/middleware"
)

// TenantFromContext retrieves the tenant from the request context
func TenantFromContext(ctx context.Context) (string, error) {
	tenant, ok := ctx.Value(TenantKey).(string)
	if !ok || tenant == "" {
		return "", mw.ErrTenantInvalid
	}
	return tenant, nil
}

// ContextWithTenant sets the tenant in the request context
func ContextWithTenant(ctx context.Context, tenant string) context.Context {
	ctx = context.WithValue(ctx, TenantKey, tenant)
	return ctx
}

// WithTenant creates new middleware which sets the tenant in the request context
func WithTenant(config WithTenantConfig) func(next http.Handler) http.Handler {
	if config.DB == nil {
		panic(mw.ErrDBInvalid.Error())
	}

	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}

	if config.TenantGetters == nil {
		config.TenantGetters = DefaultTenantGetters
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
				http.Error(w, mw.ErrTenantInvalid.Error(), http.StatusBadRequest)
				return
			}
			// set tenant in context
			r = r.WithContext(ContextWithTenant(r.Context(), tenant))
			next.ServeHTTP(w, r)
		})
	}
}
