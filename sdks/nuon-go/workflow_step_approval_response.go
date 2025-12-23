package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
	"github.com/nuonco/nuon-go/models"
)

func (c *client) CreateWorkflowStepApprovalResponse(
	ctx context.Context,
	workflowID string,
	workflowStepID string,
	workflowApprovalID string,
	req *models.ServiceCreateWorkflowStepApprovalResponseRequest,
) (*models.ServiceCreateWorkflowStepApprovalResponseResponse, error) {
	resp, err := c.genClient.Operations.CreateWorkflowStepApprovalResponse(&operations.CreateWorkflowStepApprovalResponseParams{
		WorkflowID: workflowID,
		ApprovalID: workflowApprovalID,
		Req:        req,
		Context:    ctx,
	},
		c.getOrgIDAuthInfo(),
	)

	if err != nil {
		return nil, err
	}

	return resp.Payload, nil

}
