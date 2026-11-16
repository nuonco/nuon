package nuon

import (
	"context"
	"net/http"
	"net/url"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateAppInputConfig(ctx context.Context, appID string, req *models.ServiceCreateAppInputConfigRequest) (*models.AppAppInputConfig, error) {
	resp, err := c.genClient.Operations.CreateAppInputConfig(&operations.CreateAppInputConfigParams{
		Req:     req,
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppInputLatestConfig(ctx context.Context, appID string) (*models.AppAppInputConfig, error) {
	resp, err := c.genClient.Operations.GetAppInputLatestConfig(&operations.GetAppInputLatestConfigParams{
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppInstallInputConfig(ctx context.Context, appID, appBranchID string) (*models.AppAppInputConfig, error) {
	requestURL := c.APIURL + "/v1/apps/" + url.PathEscape(appID) + "/install-input-config"
	if appBranchID != "" {
		requestURL += "?" + url.Values{"app_branch_id": {appBranchID}}.Encode()
	}

	var cfg models.AppAppInputConfig
	if err := c.triggerRequest(ctx, http.MethodGet, requestURL, nil, http.StatusOK, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func (c *client) GetAppInputConfig(ctx context.Context, appID, appInputConfigID string) (*models.AppAppInputConfig, error) {
	resp, err := c.genClient.Operations.GetAppInputConfig(&operations.GetAppInputConfigParams{
		AppID:         appID,
		InputConfigID: appInputConfigID,
		Context:       ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppInputConfigs(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppAppInputConfig, bool, error) {
	params := &operations.GetAppInputConfigsParams{
		AppID:   appID,
		Context: ctx,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.GetAppInputConfigsReader{})
	resp, err := c.genClient.Operations.GetAppInputConfigs(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}
