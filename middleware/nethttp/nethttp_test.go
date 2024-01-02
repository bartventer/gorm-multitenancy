package nethttp

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bartventer/gorm-multitenancy/v2/internal"
	mw "github.com/bartventer/gorm-multitenancy/v2/middleware"
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
			name: "test-with-tenant-search-path",
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
			defer func() {
				if r := recover(); r != nil {
					if tt.wantErr && r != mw.ErrDBInvalid.Error() {
						t.Errorf("The code did not panic as expected. Got %v", r)
					}
				}
			}()
			// setup the middleware
			middleware := WithTenant(tt.args.config)
			// setup the handler
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// tenant from context, should be same as tenant from search path
				tenant, err := TenantFromContext(r.Context())
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				if tenant != tt.want {
					http.Error(w, fmt.Sprintf("expected tenant %s, got %s", tt.want, tenant), http.StatusInternalServerError)
					return
				}
				fmt.Fprint(w, tenant)
			}))

			// Create a request to pass to our handler.
			req, err := http.NewRequest("GET", "/", nil)
			if err != nil {
				t.Fatal(err)
			}
			req.Host = tt.args.tenant + ".example.com"

			// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			// Check the status code is what we expect.
			assert.Equal(t, http.StatusOK, rr.Code)

			// Check the response body is what we expect.
			assert.Equal(t, tt.want, rr.Body.String())

		})
	}
}

func TestTenantFromContext(t *testing.T) {
	type args struct {
		ctx context.Context
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
				ctx: ContextWithTenant(context.Background(), "tenant1"),
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "invalid",
			args: args{
				ctx: context.Background(),
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
