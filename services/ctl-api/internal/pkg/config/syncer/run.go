package syncer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/lib/pq"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/config"
	configsync "github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	actionshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/actions/helpers"
	appshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	componenthelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/components/helpers"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	runbookshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/runbooks/helpers"
	vcshelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/vcs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/config/syncer/triggers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/terraform"
)

// RunDeps carries everything the syncer needs to run against the database.
type RunDeps struct {
	DB               *gorm.DB
	AppsHelpers      *appshelpers.Helpers
	ComponentHelpers *componenthelpers.Helpers
	ActionsHelpers   *actionshelpers.Helpers
	RunbooksHelpers  *runbookshelpers.Helpers
	InstallHelpers   *installhelpers.Helpers
	VCSHelpers       *vcshelpers.Helpers
	TFClient         terraform.Client
}

type RunRequest struct {
	AppID       string
	AppConfigID string

	// DispatchBuilds: see components.SyncComponentParams.DispatchBuilds.
	DispatchBuilds bool
}

type RunResult struct {
	AppConfigID         string
	ComponentIDs        []string
	ActionIDs           []string
	RunbookIDs          []string
	ComponentsScheduled []configsync.ComponentState

	// Caller must provision these queues after Run returns; the sync is transactional.
	ComponentsCreated []string
}

// Run syncs an app config from its stored intermediate config, driving the
// status through syncing to active or error. The single path from intermediate
// config to database records — the branch run and POST /configs/:id/sync both
// go through it.
func Run(ctx context.Context, deps RunDeps, req RunRequest) (*RunResult, error) {
	var appConfig app.AppConfig
	if res := deps.DB.WithContext(ctx).First(&appConfig, "id = ?", req.AppConfigID); res.Error != nil {
		return nil, fmt.Errorf("unable to load app config: %w", res.Error)
	}

	setStatus(ctx, deps.DB, &appConfig, app.AppConfigStatusSyncing, "syncing config")

	intermediateJSON, err := appConfig.IntermediateConfig.Get(ctx)
	if err != nil {
		return nil, markSyncFailed(ctx, deps.DB, &appConfig, fmt.Errorf("unable to get intermediate config: %w", err))
	}

	var cfg config.AppConfig
	decoder := json.NewDecoder(strings.NewReader(intermediateJSON))
	decoder.UseNumber()
	if err := decoder.Decode(&cfg); err != nil {
		return nil, markSyncFailed(ctx, deps.DB, &appConfig, fmt.Errorf("unable to unmarshal intermediate config: %w", err))
	}

	var opts []Option
	if req.DispatchBuilds {
		opts = append(opts, WithComponentBuildDispatch())
	}

	var result RunResult
	err = deps.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		s := NewDBSyncer(
			tx,
			deps.AppsHelpers,
			deps.ComponentHelpers,
			deps.ActionsHelpers,
			deps.RunbooksHelpers,
			deps.InstallHelpers,
			deps.VCSHelpers,
			deps.TFClient,
			req.AppID,
			&cfg,
			req.AppConfigID,
			opts...,
		)
		if err := s.Sync(ctx); err != nil {
			return err
		}
		activeStatus := app.NewCompositeStatus(ctx, app.Status(app.AppConfigStatusActive))
		activeStatus.StatusHumanDescription = "synced successfully"
		if err := triggers.Sync(
			ctx, tx, &cfg, appConfig.OrgID, appConfig.AppID, appConfig.ID,
		); err != nil {
			return fmt.Errorf("unable to sync triggers: %w", err)
		}

		if err = tx.Model(&appConfig).Updates(map[string]any{
			"status":             app.AppConfigStatusActive,
			"status_description": "synced successfully",
			"status_v2":          activeStatus,
			"component_ids":      pq.StringArray(s.GetComponentStateIds()),
			"action_ids":         pq.StringArray(s.GetActionStateIds()),
			"runbook_ids":        pq.StringArray(s.GetRunbookStateIds()),
		}).Error; err != nil {
			return errors.Wrap(err, "unable to activate app config")
		}

		result = RunResult{
			AppConfigID:         s.GetAppConfigID(),
			ComponentIDs:        s.GetComponentStateIds(),
			ActionIDs:           s.GetActionStateIds(),
			RunbookIDs:          s.GetRunbookStateIds(),
			ComponentsScheduled: s.GetComponentsScheduled(),
			ComponentsCreated:   s.GetComponentsCreated(),
		}

		return nil
	})
	if err != nil {
		return nil, markSyncFailed(ctx, deps.DB, &appConfig, err)
	}

	return &result, nil
}

func setStatus(ctx context.Context, db *gorm.DB, appConfig *app.AppConfig, status app.AppConfigStatus, description string) {
	statusV2 := app.NewCompositeStatus(ctx, app.Status(status))
	statusV2.StatusHumanDescription = description
	db.WithContext(ctx).Model(appConfig).Updates(map[string]any{
		"status":             status,
		"status_description": description,
		"status_v2":          statusV2,
	})
}

// markSyncFailed records the failure on the config and returns err unchanged.
func markSyncFailed(ctx context.Context, db *gorm.DB, appConfig *app.AppConfig, err error) error {
	setStatus(ctx, db, appConfig, app.AppConfigStatusError, fmt.Sprintf("sync failed: %s", signal.HumanError(err)))
	return err
}
