package nethttp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	multitenancy "github.com/bartventer/gorm-multitenancy/v7"
)

func ExampleDefaultTenantFromSubdomain() {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, "http://test.domain.com", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set the host of the request
	req.Host = "test.domain.com"

	// Extract the subdomain from the request
	subdomain, err := DefaultTenantFromSubdomain(req)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Subdomain:", subdomain)
	}

	// Output:
	// Subdomain: test
}

func ExampleDefaultTenantFromHeader() {
	// Create a new HTTP request
	req, err := http.NewRequest(http.MethodGet, "http://test.domain.com", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set the XTenantHeader of the request
	req.Header.Set(XTenantHeader, "test-tenant")

	// Extract the tenant from the request
	tenant, err := DefaultTenantFromHeader(req)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Tenant:", tenant)
	}

	// Output:
	// Tenant: test-tenant
}

func ExampleWithTenant() {
	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tenant := r.Context().Value(multitenancy.TenantKey).(string)
		fmt.Println("Tenant:", tenant)
	})

	handler := WithTenant(DefaultWithTenantConfig)(mux)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "tenant.example.com"
	rec := httptest.NewRecorder()

	// Execute the request
	handler.ServeHTTP(rec, req)

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
				tenant, ok := r.Context().Value(multitenancy.TenantKey).(string)
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
			req, err := http.NewRequest(http.MethodGet, "/", nil)
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

func TestDefaultTenantFromSubdomain(t *testing.T) {
	type args struct {
		r *http.Request
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "valid:test-from-subdomain",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   "tenant1.example.com",
					URL:    &url.URL{Path: "/"},
					TLS:    &tls.ConnectionState{}, // Simulate an https request
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "invalid:host parts < 2",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   "invalid",
					URL:    &url.URL{Path: "/"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid:host empty",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   ".invalid",
					URL:    &url.URL{Path: "/"},
				},
			},
			wantErr: true,
		},
		{
			name: "valid:test-from-subdomain-with-port",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   "tenant1.example.com:8080",
					URL:    &url.URL{Path: "/"},
					TLS:    &tls.ConnectionState{}, // Simulate an https request
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "valid:test-from-subdomain-with-multiple-subdomains",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   "tenant1.sub.example.com",
					URL:    &url.URL{Path: "/"},
					TLS:    &tls.ConnectionState{}, // Simulate an https request
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "invalid:no-subdomain",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   "example.com",
					URL:    &url.URL{Path: "/"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid:host-is-ip",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Header: map[string][]string{},
					Host:   "192.168.1.1",
					URL:    &url.URL{Path: "/"},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultTenantFromSubdomain(tt.args.r)
			if (err != nil) != tt.wantErr {
				t.Errorf("TenantFromSubdomain() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TenantFromSubdomain() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestDefaultTenantFromHeader(t *testing.T) {
	type args struct {
		r *http.Request
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
				r: &http.Request{
					Method: http.MethodGet,
					Host:   "example.com",
					Header: map[string][]string{
						"X-Tenant": {"tenant1"},
					},
					RequestURI: "/",
				},
			},
			want:    "tenant1",
			wantErr: false,
		},
		{
			name: "invalid:header empty",
			args: args{
				r: &http.Request{
					Method: http.MethodGet,
					Host:   "example.com",
					Header: map[string][]string{
						"X-Tenant": {" "},
					},
					RequestURI: "/",
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := DefaultTenantFromHeader(tt.args.r)
			if err != nil && !tt.wantErr {
				t.Errorf("TenantFromHeader() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("TenantFromHeader() = %v, want %v", got, tt.want)
			}
		})
	}
}
