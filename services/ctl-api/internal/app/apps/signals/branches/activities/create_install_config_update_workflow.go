package activities

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
)

type CreateInstallAppConfigVersionWorkflowInput struct {
	InstallID      string       `json:"install_id"`
	NewAppConfigID string       `json:"new_app_config_id"`
	AppBranchRunID string       `json:"app_branch_run_id"`
	InstallGroupID string       `json:"install_group_id"`
	PlanOnly       bool         `json:"plan_only"`
	Callback       callback.Ref `json:"callback,omitempty"`
}

type CreateInstallAppConfigVersionWorkflowOutput struct {
	WorkflowID                string `json:"workflow_id"`
	InstallAppConfigVersionID string `json:"install_config_update_id"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) CreateInstallAppConfigVersionWorkflow(ctx context.Context, input *CreateInstallAppConfigVersionWorkflowInput) (*CreateInstallAppConfigVersionWorkflowOutput, error) {
	if err := a.supersedePriorInstallUpdate(ctx, input); err != nil {
		return nil, err
	}
	update, err := a.installHelpers.CreateAppBranchConfigUpdateWorkflow(ctx, installhelpers.AppBranchConfigUpdateInput{
		InstallID:      input.InstallID,
		NewAppConfigID: input.NewAppConfigID,
		AppBranchRunID: input.AppBranchRunID,
		InstallGroupID: input.InstallGroupID,
		PlanOnly:       input.PlanOnly,
		Callback:       input.Callback,
	})
	if err != nil {
		return nil, err
	}

	return &CreateInstallAppConfigVersionWorkflowOutput{
		WorkflowID:                update.WorkflowID,
		InstallAppConfigVersionID: update.InstallAppConfigVersionID,
	}, nil
}

func (a *Activities) supersedePriorInstallUpdate(ctx context.Context, input *CreateInstallAppConfigVersionWorkflowInput) error {
	// A first provision has no prior app-branch update, but it still holds the
	// install workflow queue. Cancel it before enqueueing the config update.
	if !input.PlanOnly {
		if err := a.cancelInFlightProvision(ctx, input.InstallID); err != nil {
			return err
		}
	}

	prior, err := a.installHelpers.PriorInFlightInstallUpdate(ctx, input.InstallID, "")
	if err != nil {
		return err
	}
	if prior == nil {
		return nil
	}
	if err := a.CancelInstallWorkflow(ctx, &CancelInstallWorkflowInput{WorkflowID: prior.WorkflowID}); err != nil {
		return err
	}
	return a.installHelpers.MarkInstallSupersededForInstall(ctx, prior.AppBranchRunID, input.InstallID, input.AppBranchRunID)
}

func (a *Activities) cancelInFlightProvision(ctx context.Context, installID string) error {
	var workflows []app.Workflow
	err := a.db.WithContext(ctx).
		Where(app.Workflow{
			OwnerID:   installID,
			OwnerType: "installs",
			Type:      app.WorkflowTypeProvision,
		}).
		Where("finished_at IS NULL").
		Where("status->>'status' NOT IN ?", []string{
			string(app.StatusError),
			string(app.StatusCancelled),
			string(app.StatusSuccess),
		}).
		Find(&workflows).Error
	if err != nil {
		return fmt.Errorf("unable to list in-flight provision workflows: %w", err)
	}
	for _, wf := range workflows {
		if err := a.CancelInstallWorkflow(ctx, &CancelInstallWorkflowInput{WorkflowID: wf.ID}); err != nil {
			return err
		}
	}

	// The cancel signal marks the version cancelled, but it may not have landed
	// yet. Clear the callback here so a late phone-home cannot complete it.
	res := a.db.WithContext(ctx).
		Model(&app.InstallStackVersion{}).
		Where("install_id = ?", installID).
		Where("status->>'status' = ?", app.InstallStackVersionStatusPendingUser).
		Select("status", "callback_ref").
		Updates(app.InstallStackVersion{
			Status:      app.NewCompositeStatus(ctx, app.StatusCancelled),
			CallbackRef: callback.Ref{},
		})
	if res.Error != nil {
		return fmt.Errorf("unable to cancel pending stack versions: %w", res.Error)
	}

	// Phone-home can flip a pending version to active before this activity runs.
	// A callback still set means the provision await has not finished.
	var latest app.InstallStackVersion
	err = a.db.WithContext(ctx).
		Where(app.InstallStackVersion{InstallID: installID}).
		Order("created_at DESC").
		First(&latest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return fmt.Errorf("unable to get latest stack version: %w", err)
	}
	if !latest.CallbackRef.IsSet() || latest.Status.Status == app.StatusCancelled {
		return nil
	}
	res = a.db.WithContext(ctx).
		Model(&app.InstallStackVersion{}).
		Where(app.InstallStackVersion{ID: latest.ID}).
		Select("status", "callback_ref").
		Updates(app.InstallStackVersion{
			Status:      app.NewCompositeStatus(ctx, app.StatusCancelled),
			CallbackRef: callback.Ref{},
		})
	if res.Error != nil {
		return fmt.Errorf("unable to cancel latest stack version: %w", res.Error)
	}
	return nil
}
