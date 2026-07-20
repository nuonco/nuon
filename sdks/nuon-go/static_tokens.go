package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateStaticToken(ctx context.Context, req *models.ServiceCreateStaticTokenRequest) (*models.GithubComNuoncoNuonServicesCtlAPIInternalAppAccountsServiceStaticTokenResponse, error) {
	resp, err := c.genClient.Operations.CreateStaticToken(&operations.CreateStaticTokenParams{
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) ListStaticTokens(ctx context.Context) ([]*models.AppToken, error) {
	resp, err := c.genClient.Operations.ListStaticTokens(&operations.ListStaticTokensParams{
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) DeleteStaticToken(ctx context.Context, tokenID string) error {
	_, err := c.genClient.Operations.DeleteStaticToken(&operations.DeleteStaticTokenParams{
		TokenID: tokenID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	return err
}
