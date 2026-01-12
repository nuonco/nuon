package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) GetInstallComponenetLastActivePlan(ctx context.Context, installID, componentID string) (*models.ServiceGetInstallComponenetLastActivePlanResponse, error) {
	resp, err := c.
		genClient.
		Operations.
		GetInstallComponenetLastActivePlan(&operations.GetInstallComponenetLastActivePlanParams{
			ComponentID: componentID,
			InstallID:   installID,
			Context:     ctx,
		}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
