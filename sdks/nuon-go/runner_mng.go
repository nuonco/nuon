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

func (c *client) RunnerMngRestart(ctx context.Context, runnerID string) error {
	_, err := c.genClient.Operations.RestartRunnerInstall(&operations.RestartRunnerInstallParams{
		RunnerID: runnerID,
		Context:  ctx,
	}, c.getOrgIDAuthInfo())
	return err
}

func (c *client) RunnerMngShutDown(ctx context.Context, runnerID string) error {
	_, err := c.genClient.Operations.ShutDownRunnerMng(&operations.ShutDownRunnerMngParams{
		RunnerID: runnerID,
		Context:  ctx,
	}, c.getOrgIDAuthInfo())
	return err
}

func (c *client) RunnerMngVMShutDown(ctx context.Context, runnerID string) error {
	_, err := c.genClient.Operations.MngVMShutDown(&operations.MngVMShutDownParams{
		RunnerID: runnerID,
		Context:  ctx,
	}, c.getOrgIDAuthInfo())
	return err
}
