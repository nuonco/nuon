package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

type CreateInstallSyncWorkflowInput struct {
	AppID                  string            `json:"app_id" validate:"required"`
	AppInstallConfigSyncID string            `json:"app_install_config_sync_id" validate:"required"`
	Metadata               map[string]string `json:"metadata"`
}

type CreateInstallSyncWorkflowOutput struct {
	WorkflowID string `json:"workflow_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateInstallSyncWorkflow(ctx context.Context, input *CreateInstallSyncWorkflowInput) (*CreateInstallSyncWorkflowOutput, error) {
	wf, err := a.helpers.CreateAppWorkflow(ctx, input.AppID, app.WorkflowTypeAppInstallSync, input.Metadata, false)
	if err != nil {
		return nil, fmt.Errorf("unable to create install sync workflow: %w", err)
	}

	if err := a.db.WithContext(ctx).
		Model(&app.AppInstallConfigSync{}).
		Where(app.AppInstallConfigSync{ID: input.AppInstallConfigSyncID}).
		Update("workflow_id", wf.ID).Error; err != nil {
		return nil, fmt.Errorf("unable to link workflow to sync: %w", err)
	}

	return &CreateInstallSyncWorkflowOutput{WorkflowID: wf.ID}, nil
}
