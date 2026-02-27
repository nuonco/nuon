package aws

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nuonco/nuon/pkg/identity"
)

const (
	// EC2 metadata service endpoints
	metadataBaseURL     = "http://169.254.169.254"
	identityDocumentURL = metadataBaseURL + "/latest/dynamic/instance-identity/document"

	// Default timeout for metadata requests
	defaultTimeout = 5 * time.Second
)

// FetcherConfig contains configuration for the AWS identity fetcher
type FetcherConfig struct {
	// Audience for the OIDC token (usually the API URL)
	Audience string

	// RunnerID to include in metadata
	RunnerID string

	// HTTPClient for making requests (optional, uses default if nil)
	HTTPClient *http.Client
}

// Fetcher fetches identity tokens from AWS EC2 instance metadata
type Fetcher struct {
	config     *FetcherConfig
	httpClient *http.Client
}

// NewFetcher creates a new AWS identity fetcher
func NewFetcher(config *FetcherConfig) *Fetcher {
	client := config.HTTPClient
	if client == nil {
		client = &http.Client{
			Timeout: defaultTimeout,
		}
	}

	return &Fetcher{
		config:     config,
		httpClient: client,
	}
}

// FetchToken fetches an identity token from AWS EC2 metadata service
// For AWS, this currently returns the instance identity document
// TODO: When AWS supports OIDC tokens in IMDS, switch to that
func (f *Fetcher) FetchToken(ctx context.Context) (*identity.Token, error) {
	// Fetch instance identity document
	req, err := http.NewRequestWithContext(ctx, "GET", identityDocumentURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// IMDSv2 requires a token, but for simplicity we'll use IMDSv1 for now
	// TODO: Add IMDSv2 support with session tokens

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch identity document: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// For now, return the identity document as the token
	// The API will need to handle this appropriately
	return &identity.Token{
		Raw:      string(body),
		Provider: identity.ProviderTypeAWS,
		RunnerID: f.config.RunnerID,
		Metadata: map[string]string{
			"audience": f.config.Audience,
		},
	}, nil
}

// Provider returns the provider type
func (f *Fetcher) Provider() identity.ProviderType {
	return identity.ProviderTypeAWS
}

// IsAvailable checks if the AWS metadata service is available
// This can be used to detect if we're running on AWS
func IsAvailable(ctx context.Context) bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", metadataBaseURL+"/latest/meta-data/", nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
