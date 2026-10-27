package internal

import "strings"

// IsNuonEmail reports whether email is an @nuon.co address. Subdomains do not match.
func IsNuonEmail(email string) bool {
	email = strings.TrimSpace(email)
	at := strings.LastIndex(email, "@")
	if at <= 0 || at == len(email)-1 {
		return false
	}
	return strings.EqualFold(email[at+1:], "nuon.co")
}
