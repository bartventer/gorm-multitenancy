// Package dsn provides utilities for parsing and formatting DSNs (Data Source Names).
package dsn

import "strings"

// StripSchemeFromURL is a helper function that strips the scheme from the provided URL string.
//
// Example:
//
//	StripSchemeFromURL("mysql://user:password@tcp(localhost:3306)/dbname") // user:password@tcp(localhost:3306)/dbname
func StripSchemeFromURL(urlstr string) string {
	schemeEndPos := strings.Index(urlstr, "://")
	if schemeEndPos != -1 {
		// Strip the scheme from the raw URL
		urlstr = urlstr[schemeEndPos+3:]
	}
	return urlstr
}
