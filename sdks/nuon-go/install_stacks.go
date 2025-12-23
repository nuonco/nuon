package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) GetInstallStack(ctx context.Context, installID string) (*models.AppInstallStack, error) {
	resp, err := c.genClient.Operations.GetInstallStackByInstallID(
		&operations.GetInstallStackByInstallIDParams{InstallID: installID, Context: ctx},
		c.getOrgIDAuthInfo(),
	)

	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}
