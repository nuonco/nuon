package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateRole(ctx context.Context, req *models.ServiceCreateRoleRequest) (*models.AppRole, error) {
	resp, err := c.genClient.Operations.CreateRole(&operations.CreateRoleParams{
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetRole(ctx context.Context, roleID string) (*models.AppRole, error) {
	resp, err := c.genClient.Operations.GetRole(&operations.GetRoleParams{
		RoleID:  roleID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateRole(ctx context.Context, roleID string, req *models.ServiceUpdateRoleRequest) (*models.AppRole, error) {
	resp, err := c.genClient.Operations.UpdateRole(&operations.UpdateRoleParams{
		RoleID:  roleID,
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) DeleteRole(ctx context.Context, roleID string) error {
	_, err := c.genClient.Operations.DeleteRole(&operations.DeleteRoleParams{
		RoleID:  roleID,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	return err
}
