package echo

import (
	"net/http"

	mw "github.com/bartventer/gorm-multitenancy/middleware"
	"github.com/labstack/echo/v4"
)

// TenantFromContext retrieves the tenant from the request context
func TenantFromContext(ctx echo.Context) (string, error) {
	tenant, ok := ctx.Get(TenantKey.String()).(string)
	if !ok || tenant == "" {
		return "", mw.ErrTenantInvalid
	}
	return tenant, nil
}

// SetTenant sets the tenant in the request context
func SetTenant(ctx echo.Context, tenant string) {
	ctx.Set(TenantKey.String(), tenant)
}

// WithTenant creates new middleware which  sets the tenant in the request context
func WithTenant(config WithTenantConfig) echo.MiddlewareFunc {
	if config.DB == nil {
		panic(mw.ErrDBInvalid.Error())
	}

	if config.Skipper == nil {
		config.Skipper = DefaultSkipper
	}

	if config.TenantGetters == nil {
		config.TenantGetters = DefaultTenantGetters
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if config.Skipper(c.Request()) {
				return next(c)
			}
			var (
				tenant string
				err    error
			)
			for _, getter := range config.TenantGetters {
				tenant, err = getter(c.Request())
				if err == nil {
					break
				}
			}
			if err != nil {
				return echo.NewHTTPError(http.StatusInternalServerError, mw.ErrTenantInvalid.Error())
			}
			// set tenant in context
			SetTenant(c, tenant)
			return next(c)

		}
	}
}
