package azure

import (
	"context"
	"errors"
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
	// Refreshable asks for a short-lived token; only set it when the caller
	// refreshes the token before it expires.
	Refreshable bool
}

var _ tokenfetcher.TokenFetcher = (*Fetcher)(nil)

func (f *Fetcher) Name() string {
	return "azure"
}

func (f *Fetcher) FetchToken(ctx context.Context, apiClient nuonrunner.Client) (*tokenfetcher.TokenFetchResult, error) {
	audience := f.requestedAudience(ctx, apiClient)
	token, err := GetIMDSToken(ctx, audience)
	var imdsErr *IMDSError
	if err != nil && audience != azure.ManagementAudience && errors.As(err, &imdsErr) && imdsErr.ResourceNotInTenant() {
		audience = azure.ManagementAudience
		token, err = GetIMDSToken(ctx, audience)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get Azure IMDS token: %w", err)
	}

	req := &models.ServiceRunnerAuthAzureRequest{
		Token:       &token,
		RunnerID:    f.RunnerID,
		Refreshable: f.Refreshable,
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
		ExpiresAt: tokenfetcher.ParseExpiresAt(resp.ExpiresAt),
		Audience:  audience,
	}, nil
}

func (f *Fetcher) requestedAudience(ctx context.Context, apiClient nuonrunner.Client) string {
	if settings, err := apiClient.GetPublicSettings(ctx, f.RunnerID); err == nil && settings != nil {
		if audience := azure.ConfiguredRunnerAudience(settings.AzureAuthAudience); audience != "" {
			return audience
		}
	}
	if audience, ok := azure.RunnerAuthResource(os.Getenv("AZURE_RUNNER_AUTH_RESOURCE")); ok {
		return audience
	}
	return azure.ManagementAudience
}
