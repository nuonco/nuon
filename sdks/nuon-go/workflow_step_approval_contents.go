package nuon

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-go/client/operations"
)

func (c *client) GetWorkflowStepApprovalContents(
	ctx context.Context,
	owner WorkflowOwner,
	workflowID string,
	workflowStepID string,
	workflowApprovalID string,
) (interface{}, error) {
	switch {
	case owner.ownedByInstall():
		resp, err := c.genClient.Operations.GetWorkflowStepApprovalContentsByInstall(&operations.GetWorkflowStepApprovalContentsByInstallParams{
			InstallID:  owner.InstallID,
			WorkflowID: workflowID,
			StepID:     workflowStepID,
			ApprovalID: workflowApprovalID,
			Context:    ctx,
		}, c.getOrgIDAuthInfo())
		if err != nil {
			return nil, err
		}
		return resp.Payload, nil
	case owner.ownedByAppBranch():
		resp, err := c.genClient.Operations.GetWorkflowStepApprovalContentsByAppBranch(&operations.GetWorkflowStepApprovalContentsByAppBranchParams{
			AppID:       owner.AppID,
			AppBranchID: owner.AppBranchID,
			WorkflowID:  workflowID,
			StepID:      workflowStepID,
			ApprovalID:  workflowApprovalID,
			Context:     ctx,
		}, c.getOrgIDAuthInfo())
		if err != nil {
			return nil, err
		}
		return resp.Payload, nil
	}

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
