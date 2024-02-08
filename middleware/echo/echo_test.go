package echo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	mw "github.com/bartventer/gorm-multitenancy/v3/middleware"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestWithTenant(t *testing.T) {
	type args struct {
		tenant string
		config WithTenantConfig
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid",
			args: args{
				tenant: "tenant1",
				config: WithTenantConfig{
					Skipper: DefaultSkipper,
					TenantGetters: []func(c echo.Context) (string, error){
						DefaultTenantFromSubdomain,
						DefaultTenantFromHeader,
					},
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "invalid: no tenant",
			args: args{
				tenant: "",
				config: WithTenantConfig{
					Skipper: DefaultSkipper,
					TenantGetters: []func(c echo.Context) (string, error){
						DefaultTenantFromSubdomain,
						DefaultTenantFromHeader,
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			rr := httptest.NewRecorder()
			c := e.NewContext(req, rr)
			defer func() {
				if r := recover(); r != nil {
					if tt.wantErr && r != mw.ErrTenantInvalid.Error() {
						t.Errorf("The code did not panic as expected. Got %v", r)
					}
				}
			}()

			h := WithTenant(tt.args.config)(func(c echo.Context) error {
				// tenant from context, should be same as tenant from search path
				tenant, ok := c.Get(tt.args.config.ContextKey.String()).(string)
				if !ok {
					return echo.NewHTTPError(http.StatusInternalServerError, "No tenant in context")
				}
				if tenant != tt.want {
					return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("tenant is not set correctly. Got %s, want %s", tenant, tt.want))
				}
				return c.String(http.StatusOK, tenant)
			})

			req.Host = tt.args.tenant + ".example.com"
			if tt.wantErr {
				assert.Error(t, h(c))
			} else {
				assert.NoError(t, h(c))
			}
		})
	}
}
