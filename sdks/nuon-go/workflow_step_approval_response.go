package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

func (c *client) CreateWorkflowStepApprovalResponse(
	ctx context.Context,
	owner WorkflowOwner,
	workflowID string,
	workflowStepID string,
	workflowApprovalID string,
	req *models.ServiceCreateWorkflowStepApprovalResponseRequest,
) (*models.ServiceCreateWorkflowStepApprovalResponseResponse, error) {
	switch {
	case owner.ownedByInstall():
		resp, err := c.genClient.Operations.CreateWorkflowStepApprovalResponseByInstall(&operations.CreateWorkflowStepApprovalResponseByInstallParams{
			InstallID:  owner.InstallID,
			WorkflowID: workflowID,
			StepID:     workflowStepID,
			ApprovalID: workflowApprovalID,
			Req:        req,
			Context:    ctx,
		}, c.getOrgIDAuthInfo())
		if err != nil {
			return nil, err
		}
		return resp.Payload, nil
	case owner.ownedByAppBranch():
		resp, err := c.genClient.Operations.CreateWorkflowStepApprovalResponseByAppBranch(&operations.CreateWorkflowStepApprovalResponseByAppBranchParams{
			AppID:       owner.AppID,
			AppBranchID: owner.AppBranchID,
			WorkflowID:  workflowID,
			StepID:      workflowStepID,
			ApprovalID:  workflowApprovalID,
			Req:         req,
			Context:     ctx,
		}, c.getOrgIDAuthInfo())
		if err != nil {
			return nil, err
		}
		return resp.Payload, nil
	}

	resp, err := c.genClient.Operations.CreateWorkflowStepApprovalResponse(&operations.CreateWorkflowStepApprovalResponseParams{
		WorkflowID: workflowID,
		StepID:     workflowStepID,
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
