package nethttp

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bartventer/gorm-multitenancy/v5/tenantcontext"
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
			name: "test-with-tenant-search-path",
			args: args{
				tenant: "tenant1",
				config: WithTenantConfig{
					Skipper: DefaultSkipper,
					TenantGetters: []func(r *http.Request) (string, error){
						DefaultTenantFromSubdomain,
						DefaultTenantFromHeader,
					},
					ContextKey: tenantcontext.TenantKey,
					ErrorHandler: func(w http.ResponseWriter, r *http.Request, _ error) {
						http.Error(w, "Invalid tenant", http.StatusInternalServerError)
					},
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// setup the middleware
			middleware := WithTenant(tt.args.config)
			// setup the handler
			handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// tenant from context, should be same as tenant from search path
				tenant, ok := r.Context().Value(tt.args.config.ContextKey).(string)
				if !ok {
					http.Error(w, "No tenant in context", http.StatusInternalServerError)
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
