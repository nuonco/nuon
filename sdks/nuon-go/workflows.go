package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type GetInstallWorkflowsQuery struct {
	Finished *bool
	Planonly *bool
	Type     string
	Search   string
	Limit    int
	Offset   int
}

func (c *client) GetInstallWorkflows(ctx context.Context, installID string, query *GetInstallWorkflowsQuery) ([]*models.AppWorkflow, bool, error) {
	params := &operations.GetWorkflowsParams{
		InstallID: installID,
		Context:   ctx,
	}

	var limit, offset int
	if query != nil {
		params.Finished = query.Finished
		params.Planonly = query.Planonly
		if query.Type != "" {
			params.Type = &query.Type
		}
		if query.Search != "" {
			params.Search = &query.Search
		}
		limit = query.Limit
		offset = query.Offset
	}
	if limit == 0 {
		limit = 10
	}
	l := int64(limit)
	o := int64(offset)
	params.Limit = &l
	params.Offset = &o

	hr := newResponseHeaderReader(&operations.GetWorkflowsReader{})
	resp, err := c.genClient.Operations.GetWorkflows(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetWorkflows(ctx context.Context, installID string, query *models.GetPaginatedQuery) ([]*models.AppWorkflow, bool, error) {
	params := &operations.GetWorkflowsParams{
		InstallID: installID,
		Context:   ctx,
	}

	params.Offset, params.Limit = applyPaginationQuery(query)

	hr := newResponseHeaderReader(&operations.GetWorkflowsReader{})
	resp, err := c.genClient.Operations.GetWorkflows(params, c.getOrgIDAuthInfo(), hr.ClientOption())
	if err != nil {
		return nil, false, err
	}

	return resp.Payload, hasNextPage(hr), nil
}

func (c *client) GetWorkflow(ctx context.Context, workflowID string) (*models.AppWorkflow, error) {
	resp, err := c.genClient.Operations.GetWorkflow(&operations.GetWorkflowParams{
		WorkflowID: workflowID,
		Context:    ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) CancelWorkflow(ctx context.Context, workflowID string) (*operations.CancelWorkflowAccepted, error) {
	resp, err := c.genClient.Operations.CancelWorkflow(&operations.CancelWorkflowParams{
		WorkflowID: workflowID,
		Context:    ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (c *client) UpdateWorkflow(ctx context.Context, workflowID string, req *models.ServiceUpdateWorkflowRequest) (*models.AppWorkflow, error) {
	resp, err := c.genClient.Operations.UpdateWorkflow(&operations.UpdateWorkflowParams{
		WorkflowID: workflowID,
		Req:        req,
		Context:    ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetWorkflowSteps(ctx context.Context, workflowID string) ([]*models.AppWorkflowStep, error) {
	resp, err := c.genClient.Operations.GetWorkflowSteps(&operations.GetWorkflowStepsParams{
		WorkflowID: workflowID,
		Context:    ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetWorkflowStep(ctx context.Context, workflowID, stepID string) (*models.AppWorkflowStep, error) {
	resp, err := c.genClient.Operations.GetWorkflowStep(&operations.GetWorkflowStepParams{
		WorkflowID: workflowID,
		StepID:     stepID,
		Context:    ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) RetryWorkflowStep(ctx context.Context, workflowID, stepID string, req *models.ServiceRetryWorkflowStepRequest) error {
	// Note: req parameter is ignored in the current API - the endpoint no longer accepts a request body
	_, err := c.genClient.Operations.RetryWorkflowStep(&operations.RetryWorkflowStepParams{
		WorkflowID: workflowID,
		StepID:     stepID,
		Context:    ctx,
	}, c.getOrgIDAuthInfo())
	return err
}
