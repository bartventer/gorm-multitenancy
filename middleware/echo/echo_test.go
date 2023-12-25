package echo

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bartventer/gorm-multitenancy/internal"
	mw "github.com/bartventer/gorm-multitenancy/middleware"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestWithTenant(t *testing.T) {
	type args struct {
		tenant string
		config mw.WithTenantConfig
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
				config: mw.WithTenantConfig{
					DB:            internal.NewTestDB(),
					Skipper:       mw.DefaultSkipper,
					TenantGetters: mw.DefaultTenantGetters,
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "invalid:db nil (panic)",
			args: args{
				tenant: "tenant1",
				config: mw.WithTenantConfig{
					Skipper:       mw.DefaultSkipper,
					TenantGetters: mw.DefaultTenantGetters,
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
					if tt.wantErr && r != mw.ErrDBInvalid.Error() {
						t.Errorf("The code did not panic as expected. Got %v", r)
					}
				}
			}()

			h := WithTenant(tt.args.config)(func(c echo.Context) error {
				// tenant from context, should be same as tenant from search path
				tenant, err := TenantFromContext(c)
				if err != nil {
					return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
				}
				if tenant != tt.want {
					return echo.NewHTTPError(http.StatusInternalServerError, fmt.Sprintf("tenant is not set correctly. Got %s, want %s", tenant, tt.want))
				}
				return c.String(http.StatusOK, tenant)
			})

			req.Host = tt.args.tenant + ".example.com"
			assert.NoError(t, h(c))

		})
	}
}

func TestTenantFromContext(t *testing.T) {
	type args struct {
		ctx echo.Context
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
				ctx: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					rr := httptest.NewRecorder()
					ctx := e.NewContext(req, rr)
					ctx.Set(TenantKey.String(), "tenant1")
					return ctx
				}(),
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "invalid tenant",
			args: args{
				ctx: func() echo.Context {
					e := echo.New()
					req := httptest.NewRequest(http.MethodGet, "/", nil)
					rr := httptest.NewRecorder()
					ctx := e.NewContext(req, rr)
					return ctx
				}(),
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TenantFromContext(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("TenantFromContext() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TenantFromContext() = %v, want %v", got, tt.want)
			}
		})
	}
}
