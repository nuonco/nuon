package nuon

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// general methods
func (c *client) GetAuthMe(ctx context.Context) (*models.ServiceAuthMeResponse, error) {
	resp, err := c.genClient.Operations.GetAuthMe(&operations.GetAuthMeParams{
		Context: ctx,
	}, c.getApiKeyAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) ValidateToken(ctx context.Context) error {
	_, err := c.genClient.Operations.ValidateToken(&operations.ValidateTokenParams{
		Context: ctx,
	}, c.getApiKeyAuthInfo())
	return err
}

func (c *client) GetCurrentUser(ctx context.Context) (*models.AppAccount, error) {
	resp, err := c.genClient.Operations.GetCurrentUser(&operations.GetCurrentUserParams{
		Context: ctx,
	}, c.getApiKeyAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetCLIConfig(ctx context.Context) (*models.ServiceCLIConfig, error) {
	endpoint := strings.TrimRight(c.APIURL, "/") + "/v1/general/cli-config"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cli-config: %w", err)
	}
	req.Header.Set("Accept", "application/json")

	resp, err := (&http.Client{Transport: c.appTransport}).Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cli-config: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 32<<10))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve cli-config: %w", err)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, newCLIConfigHTTPError(resp.StatusCode, body)
	}

	cfg := new(models.ServiceCLIConfig)
	if err := json.Unmarshal(body, cfg); err != nil {
		return nil, fmt.Errorf("failed to retrieve cli-config: %w", err)
	}
	return cfg, nil
}

func (c *client) GetCloudPlatformRegions(ctx context.Context, cloudPlatform string) ([]*models.AppCloudPlatformRegion, error) {
	resp, err := c.genClient.Operations.GetCloudPlatformRegions(&operations.GetCloudPlatformRegionsParams{
		Context:       ctx,
		CloudPlatform: cloudPlatform,
	})
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
