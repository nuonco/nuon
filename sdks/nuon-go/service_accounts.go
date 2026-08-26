package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) ListRoles(ctx context.Context) ([]*models.AppRole, error) {
	resp, err := c.genClient.Operations.ListRoles(&operations.ListRolesParams{
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) ListServiceAccounts(ctx context.Context, includeRunners bool, query *models.GetPaginatedQuery) ([]*models.AppAccount, bool, error) {
	params := &operations.ListServiceAccountsParams{
		Context:        ctx,
		IncludeRunners: &includeRunners,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.ListServiceAccountsReader{})
	resp, err := c.genClient.Operations.ListServiceAccounts(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) CreateServiceAccount(ctx context.Context, req *models.ServiceCreateServiceAccountRequest) (*models.AppAccount, error) {
	resp, err := c.genClient.Operations.CreateServiceAccount(&operations.CreateServiceAccountParams{
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateServiceAccount(ctx context.Context, accountID string, req *models.ServiceUpdateServiceAccountRequest) (*models.AppAccount, error) {
	resp, err := c.genClient.Operations.UpdateServiceAccount(&operations.UpdateServiceAccountParams{
		AccountID: accountID,
		Req:       req,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateServiceAccountRole(ctx context.Context, accountID string, req *models.ServiceUpdateServiceAccountRoleRequest) (*models.AppAccount, error) {
	resp, err := c.genClient.Operations.UpdateServiceAccountRole(&operations.UpdateServiceAccountRoleParams{
		AccountID: accountID,
		Req:       req,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) DeleteServiceAccount(ctx context.Context, accountID string) error {
	_, err := c.genClient.Operations.DeleteServiceAccount(&operations.DeleteServiceAccountParams{
		AccountID: accountID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	return err
}

func (c *client) CreateServiceAccountToken(ctx context.Context, accountID string, req *models.ServiceCreateServiceAccountTokenRequest) (*models.ServiceCreateServiceAccountTokenResponse, error) {
	resp, err := c.genClient.Operations.CreateServiceAccountToken(&operations.CreateServiceAccountTokenParams{
		AccountID: accountID,
		Req:       req,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
