package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

type GetOrgWorkflowsQuery struct {
	Finished bool
	Planonly bool
	Limit    int64
	Offset   int64
}

func (c *client) GetOrgWorkflows(ctx context.Context, query *GetOrgWorkflowsQuery) ([]*models.AppWorkflow, error) {
	params := &operations.GetOrgWorkflowsParams{
		Context: ctx,
	}

	if query != nil {
		params.Finished = &query.Finished
		params.Planonly = &query.Planonly
		if query.Limit > 0 {
			params.Limit = &query.Limit
		}
		if query.Offset > 0 {
			params.Offset = &query.Offset
		}
	}

	resp, err := c.genClient.Operations.GetOrgWorkflows(params, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetOrgPendingApprovals(ctx context.Context) ([]*models.AppWorkflowStepApproval, error) {
	limit := int64(50)
	resp, err := c.genClient.Operations.GetOrgPendingApprovals(&operations.GetOrgPendingApprovalsParams{
		Limit:   &limit,
		Context: ctx,
	}, c.getOrgIDAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
