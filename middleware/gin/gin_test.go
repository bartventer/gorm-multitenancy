package ginmiddleware

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/gin-gonic/gin"
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
	r := gin.Default()

	r.Use(WithTenant(DefaultWithTenantConfig))

	r.GET("/", func(c *gin.Context) {
		tenant := c.GetString(TenantKey.String())
		fmt.Println("Tenant:", tenant)
		c.String(http.StatusOK, "Hello, "+tenant)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "tenant.example.com"
	w := httptest.NewRecorder()

	// Execute the request
	r.ServeHTTP(w, req)

	// Output: Tenant: tenant
}

func TestWithTenant(t *testing.T) {
	gin.SetMode(gin.TestMode)

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
					Skipper: func(c *gin.Context) bool {
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
					TenantGetters: []func(c *gin.Context) (string, error){
						func(c *gin.Context) (string, error) {
							return "", errors.New("forced error")
						},
					},
					ErrorHandler: func(c *gin.Context, err error) {
						c.JSON(http.StatusInternalServerError, gin.H{"error": "forced error"})
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
					TenantGetters: []func(c *gin.Context) (string, error){
						func(c *gin.Context) (string, error) {
							return "tenant1", nil
						},
					},
					SuccessHandler: func(c *gin.Context) {
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
			r := gin.New()
			r.Use(WithTenant(tt.args.config))
			r.GET("/", func(c *gin.Context) {
				tenant := c.GetString(TenantKey.String())
				if tt.args.config.Skipper != nil && tt.args.config.Skipper(c) {
					c.String(http.StatusOK, "")
					return
				}

				if tt.args.config.SuccessHandler == nil && tenant != tt.want {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "expected tenant " + tt.want + ", got " + tenant})
					return
				}
				c.String(http.StatusOK, tenant)
			})

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Host = tt.args.tenant + ".example.com"
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			if tt.wantErr {
				assertEqual(t, http.StatusInternalServerError, w.Code)
				return
			}

			if w.Code != http.StatusOK {
				t.Errorf("Handler error = %v, wantErr %v", w.Code, tt.wantErr)
				return
			}

			assertEqual(t, tt.want, w.Body.String())
		})
	}
}
