package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) SetRunnerID(runnerID string) {
	c.RunnerID = runnerID
}

func (c *client) SetAuthToken(token string) {
	c.APIToken = token
	if c.appTransport != nil {
		c.appTransport.authToken = token
	}
}

func (c *client) GetSettings(ctx context.Context) (*models.AppRunnerGroupSettings, error) {
	resp, err := c.genClient.Operations.GetRunnerSettings(&operations.GetRunnerSettingsParams{
		RunnerID: c.RunnerID,
		Context:  ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
