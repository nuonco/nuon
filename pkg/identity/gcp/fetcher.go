package gcp

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/nuonco/nuon/pkg/identity"
)

const (
	// GCP metadata service endpoints
	metadataBaseURL = "http://metadata.google.internal/computeMetadata/v1"
	identityURL     = metadataBaseURL + "/instance/service-accounts/default/identity"

	// Default timeout for metadata requests
	defaultTimeout = 5 * time.Second
)

// FetcherConfig contains configuration for the GCP identity fetcher
type FetcherConfig struct {
	// Audience for the OIDC token (usually the API URL)
	Audience string

	// RunnerID to include in metadata
	RunnerID string

	// HTTPClient for making requests (optional, uses default if nil)
	HTTPClient *http.Client
}

// Fetcher fetches identity tokens from GCP instance metadata
type Fetcher struct {
	config     *FetcherConfig
	httpClient *http.Client
}

// NewFetcher creates a new GCP identity fetcher
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

// FetchToken fetches an identity token from GCP metadata service
func (f *Fetcher) FetchToken(ctx context.Context) (*identity.Token, error) {
	// Build URL with audience parameter
	url := fmt.Sprintf("%s?audience=%s&format=full", identityURL, f.config.Audience)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// GCP metadata service requires this header
	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch identity token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	token, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	return &identity.Token{
		Raw:      string(token),
		Provider: identity.ProviderTypeGCP,
		RunnerID: f.config.RunnerID,
		Metadata: map[string]string{
			"audience": f.config.Audience,
		},
	}, nil
}

// Provider returns the provider type
func (f *Fetcher) Provider() identity.ProviderType {
	return identity.ProviderTypeGCP
}

// IsAvailable checks if the GCP metadata service is available
// This can be used to detect if we're running on GCP
func IsAvailable(ctx context.Context) bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", metadataBaseURL, nil)
	if err != nil {
		return false
	}

	req.Header.Set("Metadata-Flavor", "Google")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// GCP returns "Google" in the Metadata-Flavor response header
	return resp.Header.Get("Metadata-Flavor") == "Google"
}
