package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) GetOrgInvites(ctx context.Context, query *models.GetPaginatedQuery) ([]*models.AppOrgInvite, bool, error) {
	params := &operations.GetOrgInvitesParams{
		Context: ctx,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.GetOrgInvitesReader{})
	resp, err := c.genClient.Operations.GetOrgInvites(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) CreateOrgInvite(ctx context.Context, req *models.ServiceCreateOrgInviteRequest) (*models.AppOrgInvite, error) {
	resp, err := c.genClient.Operations.CreateOrgInvite(&operations.CreateOrgInviteParams{
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
