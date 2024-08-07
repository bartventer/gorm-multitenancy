package echo

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/labstack/echo/v4"
)

func assertEqual[T any](t *testing.T, expected, actual T) bool {
	t.Helper()
	if !reflect.DeepEqual(expected, actual) {
		t.Errorf("expected: %v, got: %v", expected, actual)
		return false
	}
	return true
}

func ExampleWithTenant() {
	e := echo.New()

	e.Use(WithTenant(DefaultWithTenantConfig))

	e.GET("/", func(c echo.Context) error {
		tenant := c.Get(TenantKey.String()).(string)
		fmt.Println("Tenant:", tenant)
		return c.String(http.StatusOK, "Hello, "+tenant)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "tenant.example.com"
	rec := httptest.NewRecorder()

	// Execute the request
	e.ServeHTTP(rec, req)

	// Output: Tenant: tenant
}

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
			name: "test-with-tenant-search-path",
			args: args{
				tenant: "tenant1",
				config: WithTenantConfig{},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "test-with-skipper",
			args: args{
				config: WithTenantConfig{
					Skipper: func(c echo.Context) bool {
						return true
					},
				},
			},
			want:    "",
			wantErr: false,
		},
		{
			name: "test-with-error-handler",
			args: args{
				config: WithTenantConfig{
					TenantGetters: []func(c echo.Context) (string, error){
						func(c echo.Context) (string, error) {
							return "", errors.New("forced error")
						},
					},
					ErrorHandler: func(c echo.Context, err error) error {
						return echo.NewHTTPError(http.StatusInternalServerError, "forced error")
					},
				},
			},
			want:    "",
			wantErr: true,
		},
		{
			name: "test-with-success-handler",
			args: args{
				tenant: "tenant1",
				config: WithTenantConfig{
					TenantGetters: []func(c echo.Context) (string, error){
						func(c echo.Context) (string, error) {
							return "tenant1", nil
						},
					},
					SuccessHandler: func(c echo.Context) {
						t.Log("success")
					},
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			middleware := WithTenant(tt.args.config)
			handler := middleware(func(c echo.Context) error {
				tenantValue := c.Get(TenantKey.String())
				tenant, _ := tenantValue.(string)

				if tt.args.config.Skipper != nil && tt.args.config.Skipper(c) {
					_, _ = c.Response().Write([]byte(""))
					return nil
				}

				if tt.args.config.SuccessHandler == nil && tenant != tt.want {
					return echo.NewHTTPError(http.StatusInternalServerError, "expected tenant "+tt.want+", got "+tenant)
				}
				_, _ = c.Response().Write([]byte(tenant))
				return nil
			})

			req := httptest.NewRequest(echo.GET, "/", nil)
			req.Host = tt.args.tenant + ".example.com"
			rec := httptest.NewRecorder()
			c := echo.New().NewContext(req, rec)

			if tt.wantErr {
				he := handler(c).(*echo.HTTPError)
				assertEqual(t, http.StatusInternalServerError, he.Code)
				assertEqual(t, "forced error", he.Message)
				return
			}

			if err := handler(c); (err != nil) != tt.wantErr {
				t.Errorf("Handler error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			assertEqual(t, http.StatusOK, rec.Code)
			assertEqual(t, tt.want, rec.Body.String())
		})
	}
}
