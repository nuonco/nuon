package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) RegisterInstall(ctx context.Context, registration *models.CustomermanagedInstallationRegistration) (*models.ServiceInstallRegistrationResponse, error) {
	ok, created, err := c.genClient.Operations.RegisterInstall(&operations.RegisterInstallParams{
		Context: ctx,
		Request: registration,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	if ok != nil {
		return ok.Payload, nil
	}
	return created.Payload, nil
}
