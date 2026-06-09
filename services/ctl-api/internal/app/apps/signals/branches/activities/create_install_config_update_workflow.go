package activities

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/flow/signals/executeflow"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type CreateInstallConfigUpdateWorkflowInput struct {
	InstallID      string       `json:"install_id"`
	NewAppConfigID string       `json:"new_app_config_id"`
	AppBranchRunID string       `json:"app_branch_run_id"`
	InstallGroupID string       `json:"install_group_id"`
	PlanOnly       bool         `json:"plan_only"`
	Callback       callback.Ref `json:"callback,omitempty"`
}

type CreateInstallConfigUpdateWorkflowOutput struct {
	WorkflowID            string `json:"workflow_id"`
	InstallConfigUpdateID string `json:"install_config_update_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateInstallConfigUpdateWorkflow(ctx context.Context, input *CreateInstallConfigUpdateWorkflowInput) (*CreateInstallConfigUpdateWorkflowOutput, error) {
	// Get the install to find its current AppConfigID
	var install app.Install
	if err := a.db.WithContext(ctx).First(&install, "id = ?", input.InstallID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	// Create the install workflow via install helpers (handles approval config, metadata, etc.)
	metadata := map[string]string{
		"new_app_config_id": input.NewAppConfigID,
		"app_branch_run_id": input.AppBranchRunID,
		"install_group_id":  input.InstallGroupID,
	}

	wf, err := a.installHelpers.CreateWorkflow(
		ctx,
		input.InstallID,
		app.WorkflowTypeAppBranchConfigUpdate,
		metadata,
		input.PlanOnly,
	)
	if err != nil {
		return nil, fmt.Errorf("unable to create install config update workflow: %w", err)
	}

	// Create the InstallConfigUpdate tracking record
	update := app.InstallConfigUpdate{
		AppBranchRunID: input.AppBranchRunID,
		InstallGroupID: input.InstallGroupID,
		InstallID:      input.InstallID,
		OldAppConfigID: install.AppConfigID,
		NewAppConfigID: input.NewAppConfigID,
		WorkflowID:     &wf.ID,
		Status:         app.NewCompositeStatus(ctx, app.StatusPending),
	}
	if err := a.db.WithContext(ctx).Create(&update).Error; err != nil {
		return nil, fmt.Errorf("unable to create install config update: %w", err)
	}

	// Enqueue the workflow for execution on the install's queue.
	queue, err := a.queueClient.GetQueueByOwner(ctx, input.InstallID, "installs")
	if err != nil {
		return nil, fmt.Errorf("unable to find queue for install %s: %w", input.InstallID, err)
	}

	if _, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID:  queue.ID,
		Signal:   &executeflow.Signal{WorkflowID: wf.ID},
		Callback: input.Callback,
	}); err != nil {
		return nil, fmt.Errorf("unable to enqueue workflow for install %s: %w", input.InstallID, err)
	}

	return &CreateInstallConfigUpdateWorkflowOutput{
		WorkflowID:            wf.ID,
		InstallConfigUpdateID: update.ID,
	}, nil
}
