package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateToken(ctx context.Context, req *models.ServiceCreateTokenRequest) (*models.GithubComNuoncoNuonServicesCtlAPIInternalAppTokensServiceCreateTokenResponse, error) {
	resp, err := c.genClient.Operations.CreateToken(&operations.CreateTokenParams{
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) ListTokens(ctx context.Context, accountID, q string, query *models.GetPaginatedQuery) ([]*models.ServiceTokenResponse, bool, error) {
	params := &operations.ListTokensParams{
		Context: ctx,
	}
	if accountID != "" {
		params.AccountID = &accountID
	}
	if q != "" {
		params.Q = &q
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.ListTokensReader{})
	resp, err := c.genClient.Operations.ListTokens(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) DeleteToken(ctx context.Context, tokenID string) error {
	_, err := c.genClient.Operations.DeleteToken(&operations.DeleteTokenParams{
		TokenID: tokenID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	return err
}
