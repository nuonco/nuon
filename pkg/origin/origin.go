package origin

import (
	"net/url"
	"strings"
)

// IsHTTPLoopback reports whether raw is an http URL whose host is localhost,
// 127.0.0.1, or ::1. https, empty, and lookalike hosts are not.
func IsHTTPLoopback(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" {
		return false
	}
	switch strings.ToLower(u.Hostname()) {
	case "localhost", "127.0.0.1", "::1":
		return true
	default:
		return false
	}
}
