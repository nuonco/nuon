package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) GetInstallCurrentInputs(ctx context.Context, installID string) (*models.AppInstallInputs, error) {
	resp, err := c.genClient.Operations.GetCurrentInstallInputs(&operations.GetCurrentInstallInputsParams{
		InstallID: installID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetInstallInputs(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppInstallInputs, bool, error) {
	params := &operations.GetInstallInputsParams{
		InstallID: appID,
		Context:   ctx,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.GetInstallInputsReader{})
	resp, err := c.genClient.Operations.GetInstallInputs(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) CreateInstallInputs(ctx context.Context, installID string, req *models.ServiceCreateInstallInputsRequest) (*models.AppInstallInputs, error) {
	resp, err := c.genClient.Operations.CreateInstallInputs(&operations.CreateInstallInputsParams{
		InstallID: installID,
		Req:       req,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateInstallInputs(ctx context.Context, installID string, req *models.ServiceUpdateInstallInputsRequest) (*models.AppInstallInputs, error) {
	resp, err := c.genClient.Operations.UpdateInstallInputs(&operations.UpdateInstallInputsParams{
		InstallID: installID,
		Req:       req,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
