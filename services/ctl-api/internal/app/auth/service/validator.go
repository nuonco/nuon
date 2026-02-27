package service

import (
	"context"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// OIDCClaims represents validated OIDC token claims
type OIDCClaims struct {
	// Subject is the subject identifier (e.g., instance ID, service account email)
	Subject string

	// Issuer is the token issuer URL
	Issuer string

	// Audience is the list of intended recipients for the token
	Audience []string

	// ExpiresAt is the expiration timestamp (Unix time)
	ExpiresAt int64

	// IssuedAt is the issuance timestamp (Unix time)
	IssuedAt int64

	// Custom contains provider-specific claims
	Custom map[string]interface{}
}

// OIDCValidator validates cloud provider OIDC tokens
type OIDCValidator interface {
	// ValidateToken validates the OIDC token and returns the validated claims
	ValidateToken(ctx context.Context, token string) (*OIDCClaims, error)

	// ValidateRunnerIdentity validates that the token claims match the runner's install configuration
	ValidateRunnerIdentity(ctx context.Context, runner *app.Runner, claims *OIDCClaims) error
}
