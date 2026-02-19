package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) GetInstallRunnerGroup(ctx context.Context, installID string) (*models.AppRunnerGroup, error) {
	resp, err := c.genClient.Operations.GetInstallRunnerGroup(&operations.GetInstallRunnerGroupParams{
		InstallID: installID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetRunnerGroupLeader(ctx context.Context, runnerGroupID string) (*models.AppRunner, error) {
	resp, err := c.genClient.Operations.GetRunnerGroupLeader(&operations.GetRunnerGroupLeaderParams{
		RunnerGroupID: runnerGroupID,
		Context:       ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateRunnerGroupLeader(ctx context.Context, runnerGroupID string, runnerID string) error {
	_, err := c.genClient.Operations.UpdateRunnerGroupLeader(&operations.UpdateRunnerGroupLeaderParams{
		RunnerGroupID: runnerGroupID,
		Request: &models.ServiceUpdateRunnerGroupLeaderRequest{
			RunnerID: runnerID,
		},
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	return err
}
