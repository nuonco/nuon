package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) GetAppConfigTemplate(ctx context.Context, appID string, typ models.ServiceAppConfigTemplateType) (*models.ServiceAppConfigTemplate, error) {
	resp, err := c.genClient.Operations.GetAppConfigTemplate(&operations.GetAppConfigTemplateParams{
		AppID:   appID,
		Type:    string(typ),
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppConfig(ctx context.Context, appID, appConfigID string, recurse *bool) (*models.AppAppConfig, error) {
	resp, err := c.genClient.Operations.GetAppConfig(&operations.GetAppConfigParams{
		AppID:       appID,
		AppConfigID: appConfigID,
		Context:     ctx,
		Recurse:     recurse,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CreateAppConfig(ctx context.Context, appID string, req *models.ServiceCreateAppConfigRequest) (*models.AppAppConfig, error) {
	resp, err := c.genClient.Operations.CreateAppConfig(&operations.CreateAppConfigParams{
		Req:     req,
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppLatestConfig(ctx context.Context, appID string) (*models.AppAppConfig, error) {
	resp, err := c.genClient.Operations.GetAppLatestConfig(&operations.GetAppLatestConfigParams{
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppConfigs(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppAppConfig, bool, error) {
	params := &operations.GetAppConfigsParams{
		AppID:   appID,
		Context: ctx,
	}

	query = handlePaginationQuery(query)

	if query != nil {
		offset := int64(query.Offset)
		limit := int64(query.Limit)
		params.Offset = &offset
		params.Limit = &limit
	}

	resp, err := c.genClient.Operations.GetAppConfigs(&operations.GetAppConfigsParams{
		AppID:   appID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
}

func (c *client) UpdateAppConfig(ctx context.Context, appID, appConfigID string, req *models.ServiceUpdateAppConfigRequest) (*models.AppAppConfig, error) {
	resp, err := c.genClient.Operations.UpdateAppConfig(&operations.UpdateAppConfigParams{
		AppID:       appID,
		AppConfigID: appConfigID,
		Req:         req,
		Context:     ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

// UpdateAppConfigInstalls update the installs using the provided app config version.
func (c *client) UpdateAppConfigInstalls(ctx context.Context, appID, appConfigID string, req *models.ServiceUpdateAppConfigInstallsRequest) error {
	resp, err := c.genClient.Operations.UpdateAppConfigInstalls(&operations.UpdateAppConfigInstallsParams{
		AppID:       appID,
		AppConfigID: appConfigID,
		Req:         req,
		Context:     ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return err
	}

	if resp.Payload != "ok" {
		return statusErr{resp.Payload}
	}

	return nil
}
