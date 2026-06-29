package nuon

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) GetAppBranches(ctx context.Context, appID string) ([]*models.AppAppBranch, error) {
	resp, err := c.genClient.Operations.GetAppBranches(&operations.GetAppBranchesParams{
		Context: ctx,
		AppID:   appID,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppBranch(ctx context.Context, appID, appBranchID string) (*models.AppAppBranch, error) {
	resp, err := c.genClient.Operations.GetAppBranch(&operations.GetAppBranchParams{
		Context:     ctx,
		AppID:       appID,
		AppBranchID: appBranchID,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CreateAppBranch(ctx context.Context, appID string, req *models.ServiceCreateAppBranchRequest) (*models.AppAppBranch, error) {
	resp, err := c.genClient.Operations.CreateAppBranch(&operations.CreateAppBranchParams{
		Context: ctx,
		AppID:   appID,
		Req:     req,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) DeleteAppBranch(ctx context.Context, appID, appBranchID string) error {
	reqURL := fmt.Sprintf("%s/v1/apps/%s/branches/%s", c.APIURL, appID, appBranchID)

	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpClient := &http.Client{Transport: c.appTransport}
	resp, err := httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

func (c *client) UpdateAppBranch(ctx context.Context, appID, appBranchID string, req *models.ServiceUpdateAppBranchRequest) (*models.AppAppBranch, error) {
	resp, err := c.genClient.Operations.UpdateAppBranch(&operations.UpdateAppBranchParams{
		Context:     ctx,
		AppID:       appID,
		AppBranchID: appBranchID,
		Req:         req,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CreateAppBranchConfig(ctx context.Context, appID, appBranchID string, req *models.ServiceCreateAppBranchConfigRequest) (*models.AppAppBranchConfig, error) {
	resp, err := c.genClient.Operations.CreateAppBranchConfig(&operations.CreateAppBranchConfigParams{
		Context:     ctx,
		AppID:       appID,
		AppBranchID: appBranchID,
		Req:         req,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppBranchLatestConfig(ctx context.Context, appID, appBranchID string) (*models.AppAppBranchConfig, error) {
	resp, err := c.genClient.Operations.GetAppBranchLatestConfig(&operations.GetAppBranchLatestConfigParams{
		Context:     ctx,
		AppID:       appID,
		AppBranchID: appBranchID,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) TriggerAppBranchRun(ctx context.Context, appID, appBranchID string, req *models.ServiceTriggerAppBranchRunRequest) (*models.AppAppBranchRun, error) {
	resp, err := c.genClient.Operations.TriggerAppBranchRun(&operations.TriggerAppBranchRunParams{
		Context:     ctx,
		AppID:       appID,
		AppBranchID: appBranchID,
		Req:         req,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppBranchRuns(ctx context.Context, appID, appBranchID string) ([]*models.AppWorkflow, error) {
	resp, err := c.genClient.Operations.GetAppBranchRuns(&operations.GetAppBranchRunsParams{
		Context:     ctx,
		AppID:       appID,
		AppBranchID: appBranchID,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppBranchRunBuilds(ctx context.Context, appID, appBranchID, runID string) ([]*models.AppComponentBuild, error) {
	resp, err := c.genClient.Operations.GetAppBranchRunBuilds(&operations.GetAppBranchRunBuildsParams{
		Context:     ctx,
		AppID:       appID,
		AppBranchID: appBranchID,
		RunID:       runID,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppBranchRunInstallGroups(ctx context.Context, appID, appBranchID, runID string) ([]*models.AppInstallConfigUpdate, error) {
	resp, err := c.genClient.Operations.GetAppBranchRunInstallGroups(&operations.GetAppBranchRunInstallGroupsParams{
		Context:     ctx,
		AppID:       appID,
		AppBranchID: appBranchID,
		RunID:       runID,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
