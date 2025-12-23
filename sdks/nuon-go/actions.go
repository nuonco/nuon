package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) GetActionWorkflows(ctx context.Context, appID string, query *models.GetPaginatedQuery) ([]*models.AppActionWorkflow, bool, error) {
	params := &operations.GetActionWorkflowsParams{
		AppID:   appID,
		Context: ctx,
	}

	query = handlePaginationQuery(query)

	if query != nil {
		offset := int64(query.Offset)
		limit := int64(query.Limit)
		params.Offset = &offset
		params.Limit = &limit
	}

	resp, err := c.genClient.Operations.GetActionWorkflows(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
}

// deprecated
func (c *client) GetActionWorkflow(ctx context.Context, actionWorkflowID string) (*models.AppActionWorkflow, error) {
	resp, err := c.genClient.Operations.GetActionWorkflow(&operations.GetActionWorkflowParams{
		ActionWorkflowID: actionWorkflowID,
		Context:          ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetAppActionWorkflow(ctx context.Context, appID, actionWorkflowID string) (*models.AppActionWorkflow, error) {
	resp, err := c.genClient.Operations.GetAppActionWorkflow(&operations.GetAppActionWorkflowParams{
		AppID:            appID,
		ActionWorkflowID: actionWorkflowID,
		Context:          ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CreateActionWorkflow(ctx context.Context, appID string, req *models.ServiceCreateAppActionWorkflowRequest) (*models.AppActionWorkflow, error) {
	resp, err := c.genClient.Operations.CreateAppActionWorkflow(&operations.CreateAppActionWorkflowParams{
		AppID:   appID,
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) UpdateActionWorkflow(ctx context.Context, actionWorkflowID string, req *models.ServiceUpdateActionWorkflowRequest) (*models.AppActionWorkflow, error) {
	resp, err := c.genClient.Operations.UpdateAppActionWorkflow(&operations.UpdateAppActionWorkflowParams{
		ActionWorkflowID: actionWorkflowID,
		Req:              req,
		Context:          ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) DeleteActionWorkflow(ctx context.Context, actionWorkflowID string) (bool, error) {
	resp, err := c.genClient.Operations.DeleteActionWorkflow(&operations.DeleteActionWorkflowParams{
		ActionWorkflowID: actionWorkflowID,
		Context:          ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return false, err
	}

	return resp.Payload, err
}

func (c *client) GetActionWorkflowConfigs(ctx context.Context, actionWorkflowID string) ([]*models.AppActionWorkflowConfig, error) {
	resp, err := c.genClient.Operations.GetActionWorkflowConfigs(&operations.GetActionWorkflowConfigsParams{
		ActionWorkflowID: actionWorkflowID,
		Context:          ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetActionWorkflowConfig(ctx context.Context, actionWorkflowConfigID string) (*models.AppActionWorkflowConfig, error) {
	resp, err := c.genClient.Operations.GetActionWorkflowConfig(&operations.GetActionWorkflowConfigParams{
		ActionWorkflowConfigID: actionWorkflowConfigID,
		Context:                ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CreateActionWorkflowConfig(ctx context.Context, actionWorkflowID string, req *models.ServiceCreateActionWorkflowConfigRequest) (*models.AppActionWorkflowConfig, error) {
	resp, err := c.genClient.Operations.CreateActionWorkflowConfig(&operations.CreateActionWorkflowConfigParams{
		ActionWorkflowID: actionWorkflowID,
		Req:              req,
		Context:          ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetInstallActionWorkflowRecentRuns(ctx context.Context, installID, actionWorkflowID string, query *models.GetPaginatedQuery) (*models.AppInstallActionWorkflow, bool, error) {
	params := &operations.GetInstallActionWorkflowRecentRunsParams{
		InstallID:        installID,
		ActionWorkflowID: actionWorkflowID,
		Context:          ctx,
	}

	query = handlePaginationQuery(query)

	if query != nil {
		offset := int64(query.Offset)
		limit := int64(query.Limit)
		params.Offset = &offset
		params.Limit = &limit
	}

	resp, err := c.genClient.Operations.GetInstallActionWorkflowRecentRuns(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		runs, hasMore := handlePagination(resp.Payload.Runs, int64(query.Offset), int64(query.Limit))
		resp.Payload.Runs = runs
		return resp.Payload, hasMore, nil
	}

	return resp.Payload, false, nil
}

func (c *client) CreateInstallActionWorkflowRun(ctx context.Context, installID string, req *models.ServiceCreateInstallActionWorkflowRunRequest) error {
	resp, err := c.genClient.Operations.CreateInstallActionWorkflowRun(&operations.CreateInstallActionWorkflowRunParams{
		InstallID: installID,
		Req:       req,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return err
	}

	if resp.Payload != "ok" {
		return statusErr{resp.Payload}
	}

	return nil
}

func (c *client) GetInstallActionWorkflowRun(ctx context.Context, installID, runID string) (*models.AppInstallActionWorkflowRun, error) {
	resp, err := c.genClient.Operations.GetInstallActionWorkflowRun(&operations.GetInstallActionWorkflowRunParams{
		InstallID: installID,
		RunID:     runID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetActionWorkflowLatestConfig(ctx context.Context, actionWorkflowID string) (*models.AppActionWorkflowConfig, error) {
	resp, err := c.genClient.Operations.GetActionWorkflowLatestConfig(&operations.GetActionWorkflowLatestConfigParams{
		ActionWorkflowID: actionWorkflowID,
		Context:          ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
