package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) GetOrg(ctx context.Context) (*models.AppOrg, error) {
	resp, err := c.genClient.Operations.GetOrg(&operations.GetOrgParams{
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetOrgs(ctx context.Context, query *models.GetPaginatedQuery) ([]*models.AppOrg, bool, error) {
	params := &operations.GetOrgsParams{
		Context: ctx,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	if query != nil && query.Q != "" {
		q := query.Q
		params.Q = &q
	}

	hr := newResponseHeaderReader(&operations.GetOrgsReader{})
	resp, err := c.genClient.Operations.GetOrgs(params, c.getApiKeyAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) DeleteOrg(ctx context.Context) (bool, error) {
	_, err := c.genClient.Operations.DeleteOrg(&operations.DeleteOrgParams{
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return false, err
	}

	return true, nil
}

func (c *client) CreateOrg(ctx context.Context, req *models.ServiceCreateOrgRequest) (*models.AppOrg, error) {
	resp, err := c.genClient.Operations.CreateOrg(&operations.CreateOrgParams{
		Req:     req,
		Context: ctx,
	}, c.getApiKeyAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateOrg(ctx context.Context, req *models.ServiceUpdateOrgRequest) (*models.AppOrg, error) {
	resp, err := c.genClient.Operations.UpdateOrg(&operations.UpdateOrgParams{
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
