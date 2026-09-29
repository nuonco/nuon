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

func IsCLI(userAgent string) bool {
	ua := strings.ToLower(userAgent)
	for _, pattern := range cliPatterns {
		if strings.Contains(ua, pattern) {
			return true
		}
	}
	return false
}
