package middleware

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const domainURLRegexPattern = `^(https?:\/\/)?([_a-zA-Z][_a-zA-Z0-9-]{2,}\.)+[a-zA-Z0-9-]+\.[a-zA-Z0-9-]+(:[0-9]+)?$`

var domainURLRegex = regexp.MustCompile(domainURLRegexPattern)

// ExtractSubdomain extracts the subdomain from a given domain URL.
// It adds the scheme if absent and checks if the URL is valid.
// If the subdomain starts with "pg_", it returns an error indicating an invalid subdomain for schema name.
// If the URL has no subdomain, it returns an error indicating that there is no subdomain.
// Otherwise, it returns the extracted subdomain.
//
// Example:
//
//	subdomain, err := middleware.ExtractSubdomain("http://test.domain.com")
//	if err != nil {
//		fmt.Println(err) // nil
//	}
//	fmt.Println(subdomain) // test
func ExtractSubdomain(domainURL string) (string, error) {
	fmt.Println("domainURL:", domainURL)
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
	fmt.Println("hostParts:", hostParts)
	if len(hostParts) > 2 {
		subdomain := strings.Join(hostParts[:len(hostParts)-2], ".")
		if strings.HasPrefix(subdomain, "pg_") {
			return "", errors.New("invalid subdomain for schema name, subdomain starts with `pg_`")
		}
		return subdomain, nil
	}

	return "", errors.New("no subdomain found")
}
