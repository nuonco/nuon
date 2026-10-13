package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) CreateTelemetryAccessToken(ctx context.Context, relayEndpoint string) (*models.ServiceCreateTelemetryAccessTokenResponse, error) {
	resp, err := c.genClient.Operations.CreateTelemetryAccessToken(&operations.CreateTelemetryAccessTokenParams{
		Context:       ctx,
		RelayEndpoint: &relayEndpoint,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CreateInstallTelemetryAccessToken(ctx context.Context, installID, relayEndpoint string) (*models.ServiceCreateInstallTelemetryAccessTokenResponse, error) {
	resp, err := c.genClient.Operations.CreateInstallTelemetryAccessToken(&operations.CreateInstallTelemetryAccessTokenParams{
		Context:       ctx,
		InstallID:     installID,
		RelayEndpoint: relayEndpoint,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetInstallTelemetryCollectorSettings(ctx context.Context, installID string) (*models.ServiceInstallTelemetryCollectorSettings, error) {
	resp, err := c.genClient.Operations.GetInstallTelemetryCollectorSettings(&operations.GetInstallTelemetryCollectorSettingsParams{
		Context:   ctx,
		InstallID: installID,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}
