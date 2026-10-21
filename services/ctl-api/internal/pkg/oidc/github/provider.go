package github

import "strings"

// IssuerURL is GitHub's multi-tenant Actions OIDC issuer: any GitHub user can
// mint tokens from it.
const IssuerURL = "https://token.actions.githubusercontent.com"

// IsIssuer reports whether issuerURL is the GitHub Actions issuer, tolerating a
// trailing slash.
func IsIssuer(issuerURL string) bool {
	return strings.TrimSuffix(issuerURL, "/") == IssuerURL
}
