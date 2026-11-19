package activities

import (
	"context"
	"errors"
	"fmt"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/updateappconfig"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
)

type RecordInstallDesiredConfigInput struct {
	InstallID      string       `json:"install_id" validate:"required"`
	NewAppConfigID string       `json:"new_app_config_id" validate:"required"`
	AppBranchRunID string       `json:"app_branch_run_id" validate:"required"`
	InstallGroupID string       `json:"install_group_id"`
	Callback       callback.Ref `json:"callback,omitempty"`
}

type RecordInstallDesiredConfigOutput struct {
	InstallAppConfigVersionID string `json:"install_config_update_id"`
}

// RecordInstallDesiredConfig records the new desired config on the install and
// enqueues the update-app-config signal that moves it, without deploying. The
// signal carries the branch run so it never creates a second config-version row.
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
func (a *Activities) RecordInstallDesiredConfig(ctx context.Context, input *RecordInstallDesiredConfigInput) (*RecordInstallDesiredConfigOutput, error) {
	var install app.Install
	if err := a.db.WithContext(ctx).First(&install, "id = ?", input.InstallID).Error; err != nil {
		return nil, fmt.Errorf("unable to get install: %w", err)
	}

	// The activity retries on transient enqueue failures after the row exists,
	// so key off (install, run) and reuse the row instead of duplicating it.
	var existing app.InstallAppConfigVersion
	runID := input.AppBranchRunID
	err := a.db.WithContext(ctx).
		Where(app.InstallAppConfigVersion{
			InstallID:      input.InstallID,
			AppBranchRunID: &runID,
		}).
		First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("unable to look up install config update: %w", err)
	}
	if err == nil {
		if err := a.enqueueDesiredConfigSignal(ctx, input); err != nil {
			return nil, err
		}
		return &RecordInstallDesiredConfigOutput{InstallAppConfigVersionID: existing.ID}, nil
	}

	diff, err := a.installHelpers.AppBranchConfigDiff(ctx, &install, input.NewAppConfigID)
	if err != nil {
		return nil, fmt.Errorf("unable to compute config diff: %w", err)
	}

	update := app.InstallAppConfigVersion{
		InstallID:      input.InstallID,
		OldAppConfigID: install.AppConfigID,
		NewAppConfigID: input.NewAppConfigID,
		Status:         app.NewCompositeStatus(ctx, app.StatusPending),
		Metadata:       map[string]string{"source": "app-branch", "rollout": "desired-only"},
	}
	update.AppBranchRunID = &input.AppBranchRunID
	if input.InstallGroupID != "" {
		update.InstallGroupID = &input.InstallGroupID
	}
	if err := installhelpers.CreateInstallAppConfigVersionRow(ctx, a.db, &update); err != nil {
		return nil, fmt.Errorf("unable to create install config update: %w", err)
	}

	if err := a.installHelpers.SaveInstallConfigDiffBlob(ctx, update.ID, diff); err != nil {
		a.l.Warn("unable to save config diff blob", zap.Error(err))
	}

	if err := a.enqueueDesiredConfigSignal(ctx, input); err != nil {
		return nil, err
	}

	return &RecordInstallDesiredConfigOutput{InstallAppConfigVersionID: update.ID}, nil
}

func (a *Activities) enqueueDesiredConfigSignal(ctx context.Context, input *RecordInstallDesiredConfigInput) error {
	queue, err := a.queueClient.GetQueueByOwnerAndName(ctx, input.InstallID, "installs", queuenames.InstallSignalsQueueName)
	if err != nil {
		return fmt.Errorf("unable to find install signals queue: %w", err)
	}
	if _, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: queue.ID,
		Signal: &updateappconfig.Signal{
			InstallID:      input.InstallID,
			NewAppConfigID: input.NewAppConfigID,
			AppBranchRunID: input.AppBranchRunID,
			InstallGroupID: input.InstallGroupID,
			TriggeredBy:    "app-branch",
			Metadata:       map[string]string{"source": "app-branch", "rollout": "desired-only"},
		},
		OwnerID:   input.InstallID,
		OwnerType: "installs",
		Callback:  input.Callback,
	}); err != nil {
		return fmt.Errorf("unable to enqueue update-app-config signal: %w", err)
	}
	return nil
}
