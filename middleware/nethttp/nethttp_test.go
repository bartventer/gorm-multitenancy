package nethttp

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bartventer/gorm-multitenancy/v5/tenantcontext"
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
				config: WithTenantConfig{},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "test-with-skipper",
			args: args{
				config: WithTenantConfig{
					Skipper: func(r *http.Request) bool {
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
					TenantGetters: []func(r *http.Request) (string, error){
						func(r *http.Request) (string, error) {
							return "", errors.New("forced error")
						},
					},
					ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
						http.Error(w, "forced error", http.StatusInternalServerError)
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
					TenantGetters: []func(r *http.Request) (string, error){
						func(r *http.Request) (string, error) {
							return "tenant1", nil
						},
					},
					SuccessHandler: func(w http.ResponseWriter, r *http.Request) {
						fmt.Fprint(w, "success: ")
					},
				},
			},
			want:    "success: tenant1",
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
				tenant, ok := r.Context().Value(tenantcontext.TenantKey).(string)
				if !ok && tt.args.config.Skipper != nil && tt.args.config.Skipper(r) {
					// If Skipper is not nil and returns true, we don't expect a tenant in the context
					fmt.Fprint(w, "")
					return
				}
				if tt.args.config.SuccessHandler == nil && tenant != tt.want {
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
			expectedStatus := http.StatusOK
			if tt.wantErr {
				expectedStatus = http.StatusInternalServerError
			}
			if status := rr.Code; status != expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, expectedStatus)
			}

			// Check the response body is what we expect.
			expectedBody := tt.want
			if tt.wantErr {
				expectedBody = "forced error\n"
			}
			if rr.Body.String() != expectedBody {
				t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expectedBody)
			}
		})
	}
}
