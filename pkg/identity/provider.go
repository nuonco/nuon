package identity

import (
	"context"
	"time"
)

// ProviderType represents the type of identity provider
type ProviderType string

const (
	ProviderTypeAWS   ProviderType = "aws"
	ProviderTypeGCP   ProviderType = "gcp"
	ProviderTypeAzure ProviderType = "azure"
	ProviderTypeLocal ProviderType = "local"
)

// String returns the string representation of the provider type
func (p ProviderType) String() string {
	return string(p)
}

// Token represents an identity token from a cloud provider
type Token struct {
	// Raw is the raw token string (JWT or presigned request)
	Raw string

	// Provider is the provider that issued the token
	Provider ProviderType

	// RunnerID is the unique identifier for the runner
	RunnerID string

	// Metadata contains provider-specific metadata
	Metadata map[string]string
}

// Claims represents validated identity claims from a token
type Claims struct {
	// Subject is the unique identifier for the identity (instance ID, etc.)
	Subject string

	// Issuer is the token issuer URL
	Issuer string

	// Audience is the intended audience for the token
	Audience []string

	// ExpiresAt is when the token expires
	ExpiresAt time.Time

	// Provider is the provider type
	Provider ProviderType

	// Custom contains provider-specific claims
	Custom map[string]interface{}
}

// ValidationContext contains information needed to validate runner identity
type ValidationContext struct {
	// RunnerID is the expected runner ID
	RunnerID string

	// InstallID is the install this runner belongs to
	InstallID string

	// OrgID is the organization this runner belongs to
	OrgID string
}

// Fetcher is the interface for fetching identity tokens from cloud providers
// This is used by runners to obtain their identity tokens
type Fetcher interface {
	// FetchToken fetches an identity token from the cloud provider
	FetchToken(ctx context.Context) (*Token, error)

	// Provider returns the provider type
	Provider() ProviderType
}

// Validator is the interface for validating identity tokens
// This is used by the API to validate runner identity tokens
type Validator interface {
	// ValidateToken validates a token and returns the claims
	ValidateToken(ctx context.Context, token string) (*Claims, error)

	// ValidateIdentity validates that the claims match the expected runner identity
	ValidateIdentity(ctx context.Context, claims *Claims, valCtx *ValidationContext) error

	// Provider returns the provider type
	Provider() ProviderType
}
