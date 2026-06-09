package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateAppSandboxConfig(ctx context.Context, appID string, req *models.ServiceCreateAppSandboxConfigRequest) (*models.AppAppSandboxConfig, error) {
	resp, err := c.genClient.Operations.CreateAppSandboxConfig(&operations.CreateAppSandboxConfigParams{
		Req:     req,
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppSandboxLatestConfig(ctx context.Context, appID string) (*models.AppAppSandboxConfig, error) {
	resp, err := c.genClient.Operations.GetAppSandboxLatestConfig(&operations.GetAppSandboxLatestConfigParams{
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppSandboxConfigs(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppAppSandboxConfig, bool, error) {
	params := &operations.GetAppSandboxConfigsParams{
		AppID:   appID,
		Context: ctx,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.GetAppSandboxConfigsReader{})
	resp, err := c.genClient.Operations.GetAppSandboxConfigs(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetAppSandboxBuilds(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppAppSandboxBuild, bool, error) {
	params := &operations.GetAppSandboxBuildsParams{
		AppID:   appID,
		Context: ctx,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.GetAppSandboxBuildsReader{})
	resp, err := c.genClient.Operations.GetAppSandboxBuilds(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}
