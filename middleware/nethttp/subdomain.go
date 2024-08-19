package nethttp

import (
	"errors"
	"fmt"
	"net"
	"strings"
)

// Errors
var (
	ErrInvalidHost      = errors.New("invalid host")
	ErrInvalidSubdomain = errors.New("invalid subdomain")
)

type wrapped struct {
	err  error
	host string
	msg  string
}

func (w *wrapped) Error() string {
	return fmt.Sprintf("%v: %s - host %q", w.err, w.msg, w.host)
}

func (w *wrapped) Unwrap() error {
	return w.err
}

// SubdomainOptions contains the configuration for extracting a subdomain.
type SubdomainOptions struct {
	DisallowedSubdomains []string // Disallowed subdomains
	DisallowedPrefixes   []string // Disallowed prefixes
}

func (o *SubdomainOptions) apply(opts ...SubdomainOption) {
	for _, opt := range opts {
		opt(o)
	}
}

// SubdomainOption is a function that configures the [SubdomainOptions].
type SubdomainOption func(*SubdomainOptions)

// WithDisallowedSubdomains sets the disallowed subdomains.
func WithDisallowedSubdomains(subdomains ...string) SubdomainOption {
	return func(o *SubdomainOptions) {
		o.DisallowedSubdomains = subdomains
	}
}

// WithDisallowedPrefixes sets the disallowed prefixes.
func WithDisallowedPrefixes(prefixes ...string) SubdomainOption {
	return func(o *SubdomainOptions) {
		o.DisallowedPrefixes = prefixes
	}
}

// isIPAddress checks if the given host is an IP address (IPv4 or IPv6).
func isIPAddress(host string) bool {
	return net.ParseIP(host) != nil
}

// ExtractSubdomain extracts the first subdomain from a host string, typically
// obtained from the Host field of a [net/http.Request.URL]. It returns an error if
// the host is an RFC 3986 IP address, if no subdomain is found, if the subdomain
// is disallowed, or if the subdomain contains a disallowed prefix.
//
// The host is expected to be in the following format:
//
//	subdomain.domain.tld[:port]
func ExtractSubdomain(host string, opts ...SubdomainOption) (subdomain string, err error) {
	// IPv6 address
	if strings.HasPrefix(host, "[") {
		return "", &wrapped{ErrInvalidHost, host, "IPv6 addresses are not allowed"}
	}

	// Strip port
	hostNoPort := host
	if portStart := strings.LastIndex(host, ":"); portStart != -1 {
		hostNoPort = host[:portStart]
	}

	// IPv4 address
	if isIPAddress(hostNoPort) {
		return "", &wrapped{ErrInvalidHost, host, "IPv4 addresses are not allowed"}
	}

	// Split host into hostParts (subdomain, domain, tld)
	hostParts := strings.SplitN(hostNoPort, ".", 3)
	if len(hostParts) < 3 {
		return "", &wrapped{ErrInvalidHost, host, "no subdomain found"}
	}
	subdomain = hostParts[0]

	options := new(SubdomainOptions)
	options.apply(opts...)

	for _, disallowed := range options.DisallowedSubdomains {
		if subdomain == disallowed {
			return "", &wrapped{ErrInvalidSubdomain, host, fmt.Sprintf("subdomain %q is disallowed", disallowed)}
		}
	}

	for _, disallowed := range options.DisallowedPrefixes {
		if strings.HasPrefix(subdomain, disallowed) {
			return "", &wrapped{ErrInvalidSubdomain, host, fmt.Sprintf("subdomain contains a disallowed prefix: %q", disallowed)}
		}
	}

	return subdomain, nil
}
