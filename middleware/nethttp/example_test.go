package nethttp_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"

	"github.com/bartventer/gorm-multitenancy/middleware/nethttp/v8"
)

func ExampleExtractSubdomain() {
	subdomains := []string{
		"foo.example.com",
		"bar.example.com:8080",
	}

	for _, subdomain := range subdomains {
		if sub, err := nethttp.ExtractSubdomain(subdomain); err != nil {
			log.Fatalf("unexpected error for host: %q", subdomain)
		} else {
			fmt.Println(sub)
		}
	}

	// Output:
	// foo
	// bar
}

func ExampleExtractSubdomain_invalid() {
	subdomains := []string{
		// No subdomain
		"example.com",
		"example.com:8080",

		// Disallowed prefix
		"pg_sub.example.com",

		// Disallowed subdomain
		"blacklisted.example.com",

		// IPv4 addresses
		"192.168.0.1",
		"192.168.0.1:8080",

		// IPv6 addresses
		"[fe80::1]",
		"[fe80::1]:8080",
	}

	for _, subdomain := range subdomains {
		if _, err := nethttp.ExtractSubdomain(
			subdomain,
			nethttp.WithDisallowedPrefixes("pg_"),
			nethttp.WithDisallowedSubdomains("blacklisted"),
		); err == nil {
			log.Fatalf("expected error for host: %q", subdomain)
		} else {
			fmt.Println(err)
		}
	}

	// Output:
	// invalid host: no subdomain found - host "example.com"
	// invalid host: no subdomain found - host "example.com:8080"
	// invalid subdomain: subdomain contains a disallowed prefix: "pg_" - host "pg_sub.example.com"
	// invalid subdomain: subdomain "blacklisted" is disallowed - host "blacklisted.example.com"
	// invalid host: IPv4 addresses are not allowed - host "192.168.0.1"
	// invalid host: IPv4 addresses are not allowed - host "192.168.0.1:8080"
	// invalid host: IPv6 addresses are not allowed - host "[fe80::1]"
	// invalid host: IPv6 addresses are not allowed - host "[fe80::1]:8080"
}

func ExampleDefaultTenantFromSubdomain() {
	req, err := http.NewRequest(http.MethodGet, "http://test.domain.com", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Host = "test.domain.com"

	subdomain, err := nethttp.DefaultTenantFromSubdomain(req)
	if err != nil {
		fmt.Println("Error:", err)
	} else {
		fmt.Println("Subdomain:", subdomain)
	}

	// Output:
	// Subdomain: test
}

func ExampleDefaultTenantFromHeader() {
	req, err := http.NewRequest(http.MethodGet, "http://test.domain.com", nil)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	req.Header.Set(nethttp.XTenantHeader, "test-tenant")

	tenant, err := nethttp.DefaultTenantFromHeader(req)
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
		tenant := r.Context().Value(nethttp.TenantKey).(string)
		fmt.Println("Tenant:", tenant)
	})

	handler := nethttp.WithTenant(nethttp.DefaultWithTenantConfig)(mux)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = "tenant.example.com"
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// Output: Tenant: tenant
}
