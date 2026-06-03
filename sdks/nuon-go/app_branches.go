package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) GetAppBranches(ctx context.Context, appID string) ([]*models.AppAppBranch, error) {
	resp, err := c.genClient.Operations.GetAppBranches(&operations.GetAppBranchesParams{
		Context: ctx,
		AppID:   appID,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppBranch(ctx context.Context, appID, appBranchID string) (*models.AppAppBranch, error) {
	resp, err := c.genClient.Operations.GetAppBranch(&operations.GetAppBranchParams{
		Context:     ctx,
		AppID:       appID,
		AppBranchID: appBranchID,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
