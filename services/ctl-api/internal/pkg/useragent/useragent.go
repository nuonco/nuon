package useragent

import "strings"

var cliPatterns = []string{
	"nuon-cli",
	"nuon/",
	"go-http-client",
	"curl",
	"wget",
	"postman",
}

// IsCLI reports whether the User-Agent indicates CLI or programmatic API usage
// rather than a browser.
func IsCLI(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	for _, pattern := range cliPatterns {
		if strings.Contains(ua, pattern) {
			return true
		}
	}
	return false
}
