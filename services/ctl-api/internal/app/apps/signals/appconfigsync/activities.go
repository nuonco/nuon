package appconfigsync

import (
	"context"
	"errors"
	"fmt"

	"go.temporal.io/sdk/temporal"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"gorm.io/gorm"

	configsync "github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	accountshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/accounts/helpers"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/syncappconfiginstalls"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	configcreated "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/configcreated"
	createdsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/created"
	updatecomponenttype "github.com/nuonco/nuon/services/ctl-api/internal/app/components/signals/updatecomponenttype"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/terraform"
)

type ActivitiesParams struct {
	fx.In

	DB               *gorm.DB `name:"psql"`
	AppsHelpers      *appshelpers.Helpers
	ComponentHelpers *componenthelpers.Helpers
	ActionsHelpers   *actionshelpers.Helpers
	RunbooksHelpers  *runbookshelpers.Helpers
	InstallHelpers   *installhelpers.Helpers
	VCSHelpers       *vcshelpers.Helpers
	QueueClient      *queueclient.Client
	TFClient         terraform.Client
	AccountsHelpers  *accountshelpers.Helpers
	L                *zap.Logger
}

type Activities struct {
	deps            syncer.RunDeps
	queueClient     *queueclient.Client
	accountsHelpers *accountshelpers.Helpers
	l               *zap.Logger
}

func NewActivities(params ActivitiesParams) *Activities {
	return &Activities{
		deps: syncer.RunDeps{
			DB:               params.DB,
			AppsHelpers:      params.AppsHelpers,
			ComponentHelpers: params.ComponentHelpers,
			ActionsHelpers:   params.ActionsHelpers,
			RunbooksHelpers:  params.RunbooksHelpers,
			InstallHelpers:   params.InstallHelpers,
			VCSHelpers:       params.VCSHelpers,
			TFClient:         params.TFClient,
		},
		queueClient:     params.QueueClient,
		accountsHelpers: params.AccountsHelpers,
		l:               params.L,
	}
}

type ApplyAppConfigInput struct {
	AppID       string `json:"app_id" validate:"required"`
	AppConfigID string `json:"app_config_id" validate:"required"`
}

type ApplyAppConfigOutput struct {
	AppConfigID  string   `json:"app_config_id"`
	ComponentIDs []string `json:"component_ids"`
	ActionIDs    []string `json:"action_ids"`
	RunbookIDs   []string `json:"runbook_ids"`

	// ComponentIDsToBuild are the components whose config changed.
	ComponentIDsToBuild []ComponentToBuild `json:"component_ids_to_build"`
}

type ComponentToBuild struct {
	ComponentID string            `json:"component_id"`
	Type        app.ComponentType `json:"type"`
}

// @temporal-gen-v2 activity
// @start-to-close-timeout 10m
// @as-wrapper
func (a *Activities) applyAppConfig(ctx context.Context, req *ApplyAppConfigInput) (*ApplyAppConfigOutput, error) {
	result, err := syncer.Run(ctx, a.deps, syncer.RunRequest{
		AppID:          req.AppID,
		AppConfigID:    req.AppConfigID,
		DispatchBuilds: true,
	})
	if err != nil {
		var syncErr configsync.SyncErr
		if errors.As(err, &syncErr) {
			return nil, temporal.NewNonRetryableApplicationError(
				fmt.Sprintf("unable to sync config: %s", err.Error()),
				"SYNC_VALIDATION_ERROR",
				err,
			)
		}
		return nil, fmt.Errorf("unable to sync config: %w", err)
	}

	if err := a.provisionCreatedComponents(ctx, result.ComponentsCreated); err != nil {
		return nil, err
	}

	toBuild := make([]ComponentToBuild, 0, len(result.ComponentsScheduled))
	for _, cmp := range result.ComponentsScheduled {
		toBuild = append(toBuild, ComponentToBuild{
			ComponentID: cmp.ID,
			Type:        app.ComponentType(cmp.Type),
		})
	}

	return &ApplyAppConfigOutput{
		AppConfigID:         result.AppConfigID,
		ComponentIDs:        result.ComponentIDs,
		ActionIDs:           result.ActionIDs,
		RunbookIDs:          result.RunbookIDs,
		ComponentIDsToBuild: toBuild,
	}, nil
}

// Runs post-commit: starting queue workflows inside the sync transaction would
// leave them behind for components a rollback removed.
func (a *Activities) provisionCreatedComponents(ctx context.Context, componentIDs []string) error {
	for _, componentID := range componentIDs {
		if _, err := a.deps.ComponentHelpers.EnsureComponentQueues(ctx, componentID); err != nil {
			return fmt.Errorf("unable to create queues for component %s: %w", componentID, err)
		}

		q, err := a.queueClient.GetQueueByOwner(ctx, componentID, "components")
		if err != nil {
			return fmt.Errorf("unable to get queue for component %s: %w", componentID, err)
		}

		dedupeKey := fmt.Sprintf("component-created:%s", componentID)
		if _, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
			QueueID:   q.ID,
			OwnerID:   componentID,
			OwnerType: "components",
			DedupeKey: &dedupeKey,
			Signal: &createdsignal.Signal{
				ComponentID: componentID,
			},
		}); err != nil {
			return fmt.Errorf("unable to enqueue created signal for component %s: %w", componentID, err)
		}
	}

	return nil
}

type FinalizeAppConfigSyncInput struct {
	AppID       string `json:"app_id" validate:"required"`
	AppConfigID string `json:"app_config_id" validate:"required"`
	AccountID   string `json:"account_id,omitempty"`
}

// finalizeAppConfigSync rolls the config out to installs and advances the
// onboarding journey. Branch-linked configs skip the rollout — the branch run
// manages their installs.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 1m
// @as-wrapper
func (a *Activities) finalizeAppConfigSync(ctx context.Context, req *FinalizeAppConfigSyncInput) error {
	var appConfig app.AppConfig
	if err := a.deps.DB.WithContext(ctx).
		Select("id", "app_id", "app_branch_id").
		First(&appConfig, "id = ?", req.AppConfigID).Error; err != nil {
		return fmt.Errorf("unable to load app config: %w", err)
	}

	if req.AccountID != "" {
		if err := a.accountsHelpers.UpdateUserJourneyStepForFirstAppSync(ctx, req.AccountID, req.AppID); err != nil {
			a.l.Warn("unable to update app_synced journey step",
				zap.String("account_id", req.AccountID),
				zap.String("app_id", req.AppID),
				zap.Error(err))
		}
	}

	if appConfig.AppBranchID.Valid && appConfig.AppBranchID.String != "" {
		return nil
	}

	q, err := a.queueClient.GetQueueByOwner(ctx, req.AppID, "apps")
	if err != nil {
		return fmt.Errorf("unable to get app queue: %w", err)
	}

	if _, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: q.ID,
		Signal: &syncappconfiginstalls.Signal{
			AppID:          req.AppID,
			NewAppConfigID: req.AppConfigID,
		},
	}); err != nil {
		return fmt.Errorf("unable to enqueue sync-app-config-installs signal: %w", err)
	}

	return nil
}

type DispatchComponentBuildsInput struct {
	Components []ComponentToBuild `json:"components" validate:"required"`
}

// dispatchComponentBuilds enqueues the same signals the per-type
// Create*ComponentConfig handlers do; config-created is what creates and
// starts the build.
//
// @temporal-gen-v2 activity
// @start-to-close-timeout 5m
// @as-wrapper
func (a *Activities) dispatchComponentBuilds(ctx context.Context, req *DispatchComponentBuildsInput) error {
	for _, cmp := range req.Components {
		q, err := a.queueClient.GetQueueByOwner(ctx, cmp.ComponentID, "components")
		if err != nil {
			return fmt.Errorf("unable to get queue for component %s: %w", cmp.ComponentID, err)
		}

		for _, sig := range []signal.Signal{
			&configcreated.Signal{ComponentID: cmp.ComponentID},
			&updatecomponenttype.Signal{ComponentID: cmp.ComponentID, ComponentType: cmp.Type},
		} {
			if _, err := a.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
				QueueID:   q.ID,
				OwnerID:   cmp.ComponentID,
				OwnerType: "components",
				Signal:    sig,
			}); err != nil {
				return fmt.Errorf("unable to enqueue %s signal for component %s: %w", sig.Type(), cmp.ComponentID, err)
			}
		}
	}

	return nil
}
