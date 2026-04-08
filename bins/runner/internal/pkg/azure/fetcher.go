package azure

import (
	"context"
	"fmt"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/tokenfetcher"
)

// Fetcher implements tokenfetcher.TokenFetcher for Azure managed identity.
type Fetcher struct {
	RunnerID string
}

var _ tokenfetcher.TokenFetcher = (*Fetcher)(nil)

func (f *Fetcher) Name() string {
	return "azure"
}

func (f *Fetcher) FetchToken(ctx context.Context, apiClient nuonrunner.Client) (*tokenfetcher.TokenFetchResult, error) {
	token, err := GetIMDSToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure IMDS token: %w", err)
	}

	req := &models.ServiceRunnerAuthAzureRequest{
		Token:    &token,
		RunnerID: f.RunnerID,
	}

	resp, err := apiClient.RunnerAuthAzure(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to authenticate with Azure: %w", err)
	}

	if !resp.Authenticated {
		return nil, fmt.Errorf("authentication failed: runner was not authenticated")
	}

	return &tokenfetcher.TokenFetchResult{
		RunnerID:  resp.RunnerID,
		AccountID: resp.TenantID,
		Token:     resp.Token,
	}, nil
}
