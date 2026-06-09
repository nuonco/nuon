package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// install deploys
func (c *client) GetInstallDeploys(ctx context.Context, installID string, query *models.GetPaginatedQuery) ([]*models.AppInstallDeploy, bool, error) {
	params := &operations.GetInstallDeploysParams{
		InstallID: installID,
		Context:   ctx,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.GetInstallDeploysReader{})
	resp, err := c.genClient.Operations.GetInstallDeploys(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) CreateInstallDeploy(ctx context.Context, installID string, req *models.ServiceCreateInstallDeployRequest) (*models.AppInstallDeploy, error) {
	resp, err := c.genClient.Operations.CreateInstallDeploy(&operations.CreateInstallDeployParams{
		InstallID: installID,
		Req:       req,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetInstallLatestDeploy(ctx context.Context, installID string) (*models.AppInstallDeploy, error) {
	resp, err := c.genClient.Operations.GetInstallLatestDeploy(&operations.GetInstallLatestDeployParams{
		InstallID: installID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetInstallDeploy(ctx context.Context, installID string, deployID string) (*models.AppInstallDeploy, error) {
	resp, err := c.genClient.Operations.GetInstallDeploy(&operations.GetInstallDeployParams{
		InstallID: installID,
		DeployID:  deployID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetRunnerJobPlan(ctx context.Context, runnerJobID string) (string, error) {
	resp, err := c.genClient.Operations.GetRunnerJobPlan(&operations.GetRunnerJobPlanParams{
		RunnerJobID: runnerJobID,
		Context:     ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return "", err
	}

	return resp.Payload, nil
}
