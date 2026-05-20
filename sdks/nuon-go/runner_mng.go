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

func (c *client) GetRunnerCardDetails(ctx context.Context, runnerID string) (*models.ServiceRunnerCardDetailsResponse, error) {
	resp, err := c.genClient.Operations.GetRunnerCardDetails(&operations.GetRunnerCardDetailsParams{
		RunnerID: runnerID,
		Context:  ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetRunnerJobs(ctx context.Context, runnerID string, groups string, limit int64) ([]*models.AppRunnerJob, error) {
	resp, err := c.genClient.Operations.GetRunnerJobs(&operations.GetRunnerJobsParams{
		RunnerID: runnerID,
		Groups:   &groups,
		Limit:    &limit,
		Context:  ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) ListRunnerProcesses(ctx context.Context, runnerID string, status string, limit int64) ([]*models.AppRunnerProcess, error) {
	resp, err := c.genClient.Operations.ListRunnerProcesses(&operations.ListRunnerProcessesParams{
		RunnerID: runnerID,
		Status:   &status,
		Limit:    &limit,
		Context:  ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetLatestRunnerHeartBeats(ctx context.Context, runnerID string) (models.ServiceLatestRunnerHeartBeats, error) {
	resp, err := c.genClient.Operations.GetLatestRunnerHeartBeat(&operations.GetLatestRunnerHeartBeatParams{
		RunnerID: runnerID,
		Context:  ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetRunnerRecentHealthChecks(ctx context.Context, runnerID string, processID string) ([]*models.AppRunnerHealthCheck, error) {
	resp, err := c.genClient.Operations.GetRunnerRecentHealthChecks(&operations.GetRunnerRecentHealthChecksParams{
		RunnerID:  runnerID,
		ProcessID: &processID,
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
