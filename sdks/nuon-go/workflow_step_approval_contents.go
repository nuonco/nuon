package nuon

import (
	"context"

	"github.com/nuonco/nuon-go/client/operations"
)

func (c *client) GetWorkflowStepApprovalContents(
	ctx context.Context,
	workflowID string,
	workflowStepID string,
	workflowApprovalID string,
) (interface{}, error) {
	resp, err := c.genClient.Operations.GetWorkflowStepApprovalContents(&operations.GetWorkflowStepApprovalContentsParams{
		WorkflowID: workflowID,
		StepID:     workflowStepID,
		ApprovalID: workflowApprovalID,
		Context:    ctx,
	},
		c.getOrgIDAuthInfo(),
	)

	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}
