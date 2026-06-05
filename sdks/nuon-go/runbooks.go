package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

// GetAppRunbook retrieves a runbook by name or ID for an app.
func (c *client) GetAppRunbook(ctx context.Context, appID, nameOrID string) (*models.AppRunbook, error) {
	resp, err := c.genClient.Operations.GetRunbook(&operations.GetRunbookParams{
		AppID:     appID,
		RunbookID: nameOrID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

// CreateRunbook creates a new runbook for an app.
func (c *client) CreateRunbook(ctx context.Context, appID string, req *models.ServiceCreateRunbookRequest) (*models.AppRunbook, error) {
	resp, err := c.genClient.Operations.CreateRunbook(&operations.CreateRunbookParams{
		AppID:   appID,
		Req:     req,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

// UpdateRunbook updates a runbook.
func (c *client) UpdateRunbook(ctx context.Context, runbookID string, req *models.ServiceUpdateRunbookRequest) (*models.AppRunbook, error) {
	resp, err := c.genClient.Operations.UpdateRunbook(&operations.UpdateRunbookParams{
		AppID:     "_",
		RunbookID: runbookID,
		Req:       req,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

// CreateRunbookConfig creates a new config for a runbook.
func (c *client) CreateRunbookConfig(ctx context.Context, runbookID string, req *models.ServiceCreateRunbookConfigRequest) (*models.AppRunbookConfig, error) {
	resp, err := c.genClient.Operations.CreateRunbookConfig(&operations.CreateRunbookConfigParams{
		AppID:     "_",
		RunbookID: runbookID,
		Req:       req,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

// GetInstallRunbooks retrieves all runbooks for an install.
func (c *client) GetInstallRunbooks(ctx context.Context, installID string) ([]*models.AppInstallRunbook, error) {
	resp, err := c.genClient.Operations.GetInstallRunbooks(&operations.GetInstallRunbooksParams{
		InstallID: installID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

// GetInstallRunbook retrieves a single install runbook.
func (c *client) GetInstallRunbook(ctx context.Context, installID, runbookID string) (*models.AppInstallRunbook, error) {
	resp, err := c.genClient.Operations.GetInstallRunbook(&operations.GetInstallRunbookParams{
		InstallID: installID,
		RunbookID: runbookID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

// GetInstallRunbookRuns lists runbook runs for an install, optionally filtered by runbook ID or name.
func (c *client) GetInstallRunbookRuns(ctx context.Context, installID, runbookIDOrName string, query *models.GetPaginatedQuery) ([]*models.AppInstallRunbookRun, bool, error) {
	params := &operations.GetInstallRunbookRunsParams{
		InstallID: installID,
		Context:   ctx,
	}
	if runbookIDOrName != "" {
		params.RunbookID = &runbookIDOrName
	}

	query = handlePaginationQuery(query)

	if query != nil {
		offset := int64(query.Offset)
		limit := int64(query.Limit)
		params.Offset = &offset
		params.Limit = &limit
	}

	resp, err := c.genClient.Operations.GetInstallRunbookRuns(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
}

// GetInstallRunbookRun retrieves a single runbook run on an install.
func (c *client) GetInstallRunbookRun(ctx context.Context, installID, runID string) (*models.AppInstallRunbookRun, error) {
	resp, err := c.genClient.Operations.GetInstallRunbookRun(&operations.GetInstallRunbookRunParams{
		InstallID: installID,
		RunID:     runID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}

// CreateInstallRunbookRun triggers a runbook run on an install.
func (c *client) CreateInstallRunbookRun(ctx context.Context, installID, runbookID string) (*models.AppInstallRunbookRun, error) {
	resp, err := c.genClient.Operations.CreateRunbookRun(&operations.CreateRunbookRunParams{
		InstallID: installID,
		RunbookID: runbookID,
		Context:   ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}
	return resp.Payload, nil
}
