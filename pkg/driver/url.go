package driver

import (
	"fmt"
	stdurl "net/url"
	"regexp"
)

// URL is a wrapper around the standard [stdurl.URL] type that includes the original URL string.
// It is designed to handle additional special cases for driver URL-like strings,
// such as custom formats used by various database drivers.
type URL struct {
	*stdurl.URL
	// raw is the original URL string before any normalization.
	raw string
	// standard is the normalized URL string after any normalization.
	standard string
}

// Parse parses the provided rawURL string and returns a new [URL] instance.
// It normalizes the URL to handle special cases specific to database driver URLs.
func (u *URL) Parse(rawURL string) (*URL, error) {
	var err error
	url := new(URL)
	url.raw = rawURL
	url.standard = normalizeURL(rawURL)
	url.URL, err = stdurl.Parse(url.standard)
	if err != nil {
		return nil, fmt.Errorf("failed to parse normalized URL: %w", err)
	}
	return url, nil
}

// Raw returns the original URL string before any normalization.
func (u *URL) Raw() string {
	return u.raw
}

// normalizeURL adjusts the provided URL string to match the standard URL format.
// It specifically handles special cases such as the @tcp(localhost:3306) format,
// converting them to a more standard URL format.
func normalizeURL(urlstr string) string {
	// Regular expression to match the @tcp(localhost:3306) format
	re := regexp.MustCompile(`@tcp\(([^)]+)\)`)
	matches := re.FindStringSubmatch(urlstr)
	// If the URL matches the @tcp format, adjust it; otherwise, return the original URL
	if len(matches) == 2 {
		// Replace the @tcp(localhost:3306) with @localhost:3306
		adjustedURL := re.ReplaceAllString(urlstr, "@"+matches[1])
		return adjustedURL
	}
	// Return the original URL if it doesn't match the @tcp pattern
	return urlstr
}

// ParseURL parses the provided rawURL string and returns a new [URL] instance.
// This function is a convenience wrapper around the [URL.Parse] method,
// allowing for direct parsing of raw URL strings into the URL type.
func ParseURL(rawURL string) (*URL, error) {
	return (&URL{}).Parse(rawURL)
}
