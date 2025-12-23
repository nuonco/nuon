package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
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

	query = handlePaginationQuery(query)

	if query != nil {
		offset := int64(query.Offset)
		limit := int64(query.Limit)
		params.Offset = &offset
		params.Limit = &limit
	}

	resp, err := c.genClient.Operations.GetInstallInputs(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
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

func (c *client) UpdateInstallInputs(ctx context.Context, installID string, req *models.ServiceUpdateInstallInputsRequest) (*models.AppInstallInputs, string, error) {
	hr := newResponseHeaderReader(&operations.UpdateInstallInputsReader{})

	resp, err := c.genClient.Operations.UpdateInstallInputs(&operations.UpdateInstallInputsParams{
		InstallID: installID,
		Req:       req,
		Context:   ctx,
	}, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, "", err
	}

	workflowID := hr.GetHeader(HeaderInstallWorkflowID)

	return resp.Payload, workflowID, nil
}
