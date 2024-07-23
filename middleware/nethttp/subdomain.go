package nethttp

import (
	"errors"
	"net"
	"strings"
)

// ExtractSubdomain extracts the first subdomain from a host as specified in
// [net/url.URL.Host]. It returns an error if the host is an RFC 3986 IP address or
// does not contain a subdomain.
//
// Valid examples:
//   - sub.example.com
//   - sub.example.com:8080
//
// Invalid domain examples:
//   - example.com
//   - example.com:8080
//   - pg_sub.example.com
//   - 192.168.0.1
//   - 192.168.0.1:8080
//   - [fe80::1]
//   - [fe80::1]:8080
func ExtractSubdomain(host string) (string, error) {
	// IPv6 address
	if strings.HasPrefix(host, "[") {
		return "", errors.New("domainURL is an IP address, not a subdomain")
	}

	// Strip port
	if portStart := strings.LastIndex(host, ":"); portStart != -1 {
		host = host[:portStart]
	}

	// IPv4 address
	if net.ParseIP(host) != nil {
		return "", errors.New("domainURL is an IP address, not a subdomain")
	}

	// Find the first dot in the domain
	firstDotIndex := strings.Index(host, ".")
	if firstDotIndex > 0 {
		// Extract the subdomain
		subdomain := host[:firstDotIndex]
		if strings.HasPrefix(subdomain, "pg_") {
			return "", errors.New("invalid subdomain for schema name, subdomain starts with `pg_`")
		}
		// Ensure there's another dot after the first one to confirm it's a subdomain
		if strings.Contains(host[firstDotIndex+1:], ".") {
			return subdomain, nil
		}
	}
	return "", errors.New("no subdomain found")
}
