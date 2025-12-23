package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) GetWorkflows(ctx context.Context, installID string, query *models.GetPaginatedQuery) ([]*models.AppWorkflow, bool, error) {
	params := &operations.GetWorkflowsParams{
		InstallID: installID,
		Context:   ctx,
	}

	query = handlePaginationQuery(query)

	if query != nil {
		offset := int64(query.Offset)
		limit := int64(query.Limit)
		params.Offset = &offset
		params.Limit = &limit
	}

	resp, err := c.genClient.Operations.GetWorkflows(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, false, err
	}

	if query != nil {
		items, hasMore := handlePagination(resp.Payload, int64(query.Offset), int64(query.Limit))
		return items, hasMore, nil
	}

	return resp.Payload, false, nil
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
