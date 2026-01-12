package nuonrunner

import (
	"context"

	"github.com/nuonco/nuon/sdks/nuon-runner-go/client/operations"
	"github.com/nuonco/nuon/sdks/nuon-runner-go/models"
)

func (c *client) UpdateInstallActionWorkflowRunStep(ctx context.Context, installID, workflowRunID, stepID string, req *models.ServiceUpdateInstallActionWorkflowRunStepRequest) (*models.AppInstallActionWorkflowRunStep, error) {
	resp, err := c.genClient.Operations.UpdateInstallActionWorkflowRunStep(&operations.UpdateInstallActionWorkflowRunStepParams{
		InstallID:     installID,
		StepID:        stepID,
		WorkflowRunID: workflowRunID,
		Req:           req,
		Context:       ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetInstallActionWorkflowRun(ctx context.Context, installID, workflowRunID string) (*models.AppInstallActionWorkflowRun, error) {
	resp, err := c.genClient.Operations.GetInstallActionWorkflowRun(&operations.GetInstallActionWorkflowRunParams{
		InstallID: installID,
		RunID:     workflowRunID,
		Context:   ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetActionWorkflowConfig(ctx context.Context, actionWorkflowID string) (*models.AppActionWorkflowConfig, error) {
	resp, err := c.genClient.Operations.GetActionWorkflowConfig(&operations.GetActionWorkflowConfigParams{
		ActionWorkflowConfigID: actionWorkflowID,
		Context:                ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

func (c *client) GetActionWorkflowLatestConfig(ctx context.Context, workflowID string) (*models.AppActionWorkflowConfig, error) {
	resp, err := c.genClient.Operations.GetActionWorkflowLatestConfig(&operations.GetActionWorkflowLatestConfigParams{
		ActionWorkflowID: workflowID,
		Context:          ctx,
	}, c.getAuthInfo())
	if err != nil {
		return nil, err
	}

	return resp.Payload, nil
}

