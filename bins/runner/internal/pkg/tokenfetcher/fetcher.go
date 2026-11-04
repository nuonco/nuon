package tokenfetcher

import (
	"context"
	"time"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

// TokenFetchResult holds the result of a cloud token fetch.
type TokenFetchResult struct {
	RunnerID   string
	InstanceID string
	AccountID  string
	ProjectID  string
	Token      string
	// ExpiresAt is zero when the control plane did not report an expiry.
	ExpiresAt time.Time
}

// TokenFetcher defines the interface for cloud-specific token fetching.
// Currently only Azure implements this interface; AWS and GCP use inline
// code paths in the fetchtoken package for backward compatibility.
type TokenFetcher interface {
	// FetchToken authenticates with the cloud provider and returns a runner API token.
	FetchToken(ctx context.Context, apiClient nuonrunner.Client) (*TokenFetchResult, error)

	// Name returns the name of the cloud platform (e.g., "azure").
	Name() string
}

// ParseExpiresAt parses an RFC 3339 expiry from the runner API, returning the
// zero time when it is missing or malformed.
func ParseExpiresAt(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	expiresAt, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return expiresAt
}
