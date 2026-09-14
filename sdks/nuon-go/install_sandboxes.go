package nuon

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

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
	var result models.AppWorkflowResponse
	path := fmt.Sprintf("%s/v1/installs/%s/deprovision-sandbox", c.APIURL, url.PathEscape(installID))
	err := c.triggerRequest(
		ctx,
		http.MethodPost,
		path,
		&models.ServiceDeprovisionInstallSandboxRequest{},
		http.StatusCreated,
		&result,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *client) ReprovisionInstallSandbox(ctx context.Context, installID string, skipComponents ...bool) (*models.AppWorkflowResponse, error) {
	skip := len(skipComponents) > 0 && skipComponents[0]
	var result models.AppWorkflowResponse
	path := fmt.Sprintf("%s/v1/installs/%s/reprovision-sandbox", c.APIURL, url.PathEscape(installID))
	err := c.triggerRequest(
		ctx,
		http.MethodPost,
		path,
		&models.ServiceReprovisionInstallSandboxRequest{
			PlanOnly:       false,
			SkipComponents: skip,
		},
		http.StatusCreated,
		&result,
	)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
