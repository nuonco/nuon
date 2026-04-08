package tokenfetcher

import (
	"context"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
)

// TokenFetchResult holds the result of a cloud token fetch.
type TokenFetchResult struct {
	RunnerID   string
	InstanceID string
	AccountID  string
	ProjectID  string
	Token      string
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
