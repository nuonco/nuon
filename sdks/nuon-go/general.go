package nuon

import (
	"context"

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
	resp, err := c.genClient.Operations.GetCLIConfig(&operations.GetCLIConfigParams{
		Context: ctx,
	})
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
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
