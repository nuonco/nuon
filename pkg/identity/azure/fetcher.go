package azure

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/nuonco/nuon/pkg/identity"
)

const (
	// Azure Instance Metadata Service endpoint
	metadataBaseURL = "http://169.254.169.254/metadata/identity/oauth2/token"

	// Default timeout for metadata requests
	defaultTimeout = 5 * time.Second
)

// FetcherConfig contains configuration for the Azure identity fetcher
type FetcherConfig struct {
	// Audience for the OIDC token (usually the API URL)
	Audience string

	// RunnerID to include in metadata
	RunnerID string

	// HTTPClient for making requests (optional, uses default if nil)
	HTTPClient *http.Client
}

// Fetcher fetches identity tokens from Azure Instance Metadata Service
type Fetcher struct {
	config     *FetcherConfig
	httpClient *http.Client
}

// tokenResponse represents the JSON response from Azure IMDS
type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    string `json:"expires_in"`
	ExpiresOn    string `json:"expires_on"`
	NotBefore    string `json:"not_before"`
	Resource     string `json:"resource"`
	TokenType    string `json:"token_type"`
}

// NewFetcher creates a new Azure identity fetcher
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

// FetchToken fetches an identity token from Azure Instance Metadata Service
func (f *Fetcher) FetchToken(ctx context.Context) (*identity.Token, error) {
	// Build URL with query parameters
	params := url.Values{}
	params.Add("api-version", "2018-02-01")
	params.Add("resource", f.config.Audience)

	tokenURL := fmt.Sprintf("%s?%s", metadataBaseURL, params.Encode())

	req, err := http.NewRequestWithContext(ctx, "GET", tokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Azure IMDS requires this header
	req.Header.Set("Metadata", "true")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch identity token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var tokenResp tokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	return &identity.Token{
		Raw:      tokenResp.AccessToken,
		Provider: identity.ProviderTypeAzure,
		RunnerID: f.config.RunnerID,
		Metadata: map[string]string{
			"audience":   f.config.Audience,
			"expires_in": tokenResp.ExpiresIn,
		},
	}, nil
}

// Provider returns the provider type
func (f *Fetcher) Provider() identity.ProviderType {
	return identity.ProviderTypeAzure
}

// IsAvailable checks if the Azure Instance Metadata Service is available
// This can be used to detect if we're running on Azure
func IsAvailable(ctx context.Context) bool {
	client := &http.Client{
		Timeout: 2 * time.Second,
	}

	req, err := http.NewRequestWithContext(ctx, "GET", "http://169.254.169.254/metadata/instance?api-version=2021-02-01", nil)
	if err != nil {
		return false
	}

	req.Header.Set("Metadata", "true")

	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
