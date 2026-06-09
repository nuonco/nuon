package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) GetInstallSandboxRuns(ctx context.Context, installID string, query *models.GetPaginatedQuery) ([]*models.AppInstallSandboxRun, bool, error) {
	params := &operations.GetInstallSandboxRunsParams{
		InstallID: installID,
		Context:   ctx,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.GetInstallSandboxRunsReader{})
	resp, err := c.genClient.Operations.GetInstallSandboxRuns(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetInstallSandboxRun(ctx context.Context, installID, runID string) (*models.AppInstallSandboxRun, error) {
	resp, err := c.genClient.Operations.GetInstallSandboxRunV2(&operations.GetInstallSandboxRunV2Params{
		InstallID: installID,
		RunID:     runID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) DeprovisionInstallSandbox(ctx context.Context, installID string) (*models.AppWorkflowResponse, error) {
	resp, err := c.genClient.Operations.DeprovisionInstallSandbox(&operations.DeprovisionInstallSandboxParams{
		InstallID: installID,
		Context:   ctx,
		// TODO: make this configurable
		Req: &models.ServiceDeprovisionInstallSandboxRequest{
			PlanOnly: false,
		},
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) ReprovisionInstallSandbox(ctx context.Context, installID string, skipComponents ...bool) (*models.AppWorkflowResponse, error) {
	resp, err := c.genClient.Operations.ReprovisionInstallSandbox(&operations.ReprovisionInstallSandboxParams{
		InstallID: installID,
		Context:   ctx,
		Req: &models.ServiceReprovisionInstallSandboxRequest{
			PlanOnly: false,
		},
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
