package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) CreateTelemetryAccessToken(ctx context.Context) (*models.ServiceCreateTelemetryAccessTokenResponse, error) {
	resp, err := c.genClient.Operations.CreateTelemetryAccessToken(&operations.CreateTelemetryAccessTokenParams{
		Context: ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
