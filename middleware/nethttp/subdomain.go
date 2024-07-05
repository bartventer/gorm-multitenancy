package nethttp

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

const domainURLRegexPattern = `^(https?:\/\/)?([_a-zA-Z][_a-zA-Z0-9-]{2,}\.)+[a-zA-Z0-9-]+\.[a-zA-Z0-9-]+(:[0-9]+)?$`

var domainURLRegex = regexp.MustCompile(domainURLRegexPattern)

// ExtractSubdomain extracts the first part of the subdomain from a given domain URL.
// If the URL has multiple subdomains, it only returns the first part.
//
// Valid domain URLs:
//   - http://sub.example.com
//   - https://sub.example.com
//   - http://sub.example.com:8080
//   - https://sub.example.com:8080
//   - sub.example.com
//   - sub.example.com:8080
func ExtractSubdomain(domainURL string) (string, error) {
	// Add scheme if absent
	if !strings.HasPrefix(domainURL, "http://") && !strings.HasPrefix(domainURL, "https://") {
		domainURL = "http://" + domainURL
	}

	// Check if URL is valid
	if !domainURLRegex.MatchString(domainURL) {
		return "", errors.New("invalid URL, URL should be in the format `http(s)://subdomain.domain.tld`")
	}

	u, err := url.Parse(domainURL)
	if err != nil {
		return "", err
	}

	// Split the hostname into parts
	hostParts := strings.Split(u.Hostname(), ".")
	if len(hostParts) > 2 {
		subdomain := hostParts[0]
		// Check if subdomain starts with "pg_"
		if strings.HasPrefix(subdomain, "pg_") {
			return "", errors.New("invalid subdomain for schema name, subdomain starts with `pg_`")
		}
		// Return the first part of the subdomain
		return subdomain, nil
	}
	// Return an error if no subdomain is found
	return "", errors.New("no subdomain found")
}
