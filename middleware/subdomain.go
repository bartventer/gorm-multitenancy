package middleware

import (
	"errors"
	"net/url"
	"regexp"
	"strings"
)

const domainURLRegexPattern = `^(https?:\/\/)?([_a-zA-Z][_a-zA-Z0-9-]{2,}\.)+[a-zA-Z0-9-]+\.[a-zA-Z0-9-]+(:[0-9]+)?$`

var domainURLRegex = regexp.MustCompile(domainURLRegexPattern)

// ExtractSubdomain extracts the first part of the subdomain from a given domain URL.
// It adds the scheme if absent and checks if the URL is valid.
// If the subdomain starts with "pg_", it returns an error indicating an invalid subdomain for schema name.
// If the URL has no subdomain, it returns an error indicating that there is no subdomain.
// Otherwise, it returns the extracted subdomain.
// If the URL has multiple subdomains, it only returns the first part.
//
// Example:
//
//	subdomain, err := middleware.ExtractSubdomain("http://test.domain.com")
//	if err != nil {
//		fmt.Println(err) // nil
//	}
//	fmt.Println(subdomain) // test
//
//	subdomain, err = middleware.ExtractSubdomain("http://test.sub.domain.com")
//	if err != nil {
//		fmt.Println(err) // nil
//	}
//	fmt.Println(subdomain) // test
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

	hostParts := strings.Split(u.Hostname(), ".")
	if len(hostParts) > 2 {
		subdomain := hostParts[0]
		if strings.HasPrefix(subdomain, "pg_") {
			return "", errors.New("invalid subdomain for schema name, subdomain starts with `pg_`")
		}
		return subdomain, nil
	}

	return "", errors.New("no subdomain found")
}
