package nuon

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

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

func (c *client) GetInstallRunbookRuns(ctx context.Context, installID, runbookIDOrName string, query *models.GetPaginatedQuery) ([]*models.AppInstallRunbookRun, bool, error) {
	params := &operations.GetInstallRunbookRunsParams{
		InstallID: installID,
		Context:   ctx,
	}
	if runbookIDOrName != "" {
		params.RunbookID = &runbookIDOrName
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.GetInstallRunbookRunsReader{})
	resp, err := c.genClient.Operations.GetInstallRunbookRuns(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

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

func (c *client) CreateInstallRunbookRun(ctx context.Context, installID, runbookID string) (*models.AppInstallRunbookRun, error) {
	var result models.AppInstallRunbookRun
	path := fmt.Sprintf(
		"%s/v1/installs/%s/runbooks/%s/runs",
		c.APIURL,
		url.PathEscape(installID),
		url.PathEscape(runbookID),
	)
	err := c.triggerRequest(ctx, http.MethodPost, path, nil, http.StatusCreated, &result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}
