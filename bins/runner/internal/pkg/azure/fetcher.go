package azure

import (
	"context"
	"fmt"
	"os"

	nuonrunner "github.com/nuonco/nuon/sdks/nuon-runner-go"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"

	"github.com/nuonco/nuon/bins/runner/internal/pkg/tokenfetcher"
	"github.com/nuonco/nuon/pkg/azure"
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
	resource, ok := azure.RunnerAuthResource(os.Getenv("AZURE_RUNNER_AUTH_RESOURCE"))
	if !ok {
		return nil, fmt.Errorf("AZURE_RUNNER_AUTH_RESOURCE is not the ARM audience or an api://<client-id> audience")
	}

	token, err := GetIMDSToken(ctx, resource)
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
