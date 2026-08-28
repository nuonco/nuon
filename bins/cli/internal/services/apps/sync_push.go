package apps

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/errs"
)

const (
	appConfigStatusActive  = "active"
	appConfigStatusError   = "error"
	defaultConfigSyncPoll  = time.Second * 2
	defaultConfigSyncLimit = time.Minute * 15
)

// createConfig uploads the parsed config in intermediate form. Nothing is
// converted to database records until something syncs it, either
// POST /configs/:id/sync or a branch run's sync app config step.
func (s *Service) createConfig(ctx context.Context, appID, version string, cfg *config.AppConfig, sourceArchive *config.SourceArchive, branchID string, planOnly bool) (*models.AppAppConfig, error) {
	intermediateJSON, err := json.Marshal(cfg)
	if err != nil {
		return nil, errs.WithUserFacing(err, "unable to serialize config")
	}
	var sourceJSON []byte
	if sourceArchive != nil {
		sourceJSON, err = json.Marshal(sourceArchive)
		if err != nil {
			return nil, errs.WithUserFacing(err, "unable to serialize authored config")
		}
	}

	appConfig, err := s.api.CreateAppConfig(ctx, appID, &models.ServiceCreateAppConfigRequest{
		Readme:                 cfg.Readme,
		CliVersion:             version,
		IntermediateConfigJSON: string(intermediateJSON),
		SourceConfigJSON:       string(sourceJSON),
		AppBranchID:            branchID,
		PlanOnly:               planOnly,
	})
	if err != nil {
		return nil, err
	}

	return appConfig, nil
}

// syncConfig asks the API to apply an already-uploaded app config and waits for
// the result. All conversion to database records is server-side
// (internal/pkg/config/syncer). This is the standalone path: it dispatches
// component builds itself, which is why a branch run must not use it.
func (s *Service) syncConfig(ctx context.Context, appID string, appConfig *models.AppAppConfig, opts SyncOptions) (*sync.State, error) {
	if _, err := s.api.SyncAppConfig(ctx, appID, appConfig.ID); err != nil {
		return nil, err
	}

	synced, err := s.waitForConfigSync(ctx, appID, appConfig.ID, opts.PrintJSON)
	if err != nil {
		return nil, err
	}

	return parseSyncState(synced.State), nil
}

// waitForConfigSync polls until the sync reaches a terminal state.
func (s *Service) waitForConfigSync(ctx context.Context, appID, appConfigID string, printJSON bool) (*models.AppAppConfig, error) {
	return s.waitForRunConfigSync(ctx, appID, appConfigID, "", printJSON)
}

// waitForRunConfigSync polls until the sync reaches a terminal state. When
// workflowID is set the sync is a step of that workflow, so a workflow that dies
// before reaching the step is reported instead of waiting out the full timeout.
func (s *Service) waitForRunConfigSync(ctx context.Context, appID, appConfigID, workflowID string, printJSON bool) (*models.AppAppConfig, error) {
	spinner := bubbles.NewSpinnerView(printJSON, s.cfg.Interactive)
	spinner.Start("syncing config")

	pollCtx, cancel := context.WithTimeout(ctx, defaultConfigSyncLimit)
	defer cancel()

	lastDescription := ""
	for {
		appConfig, err := s.api.GetAppConfig(pollCtx, appID, appConfigID, nil)
		if err != nil && !nuon.IsNotFound(err) && !nuon.IsServerError(err) {
			spinner.Fail(err)
			return nil, err
		}

		if appConfig != nil {
			if appConfig.StatusDescription != "" && appConfig.StatusDescription != lastDescription {
				lastDescription = appConfig.StatusDescription
				spinner.Update(lastDescription)
			}

			switch appConfig.Status {
			case appConfigStatusActive:
				spinner.Success("config synced")
				return appConfig, nil
			case appConfigStatusError:
				err := errs.NewUserFacing("%s", syncFailureMessage(appConfig))
				spinner.Fail(err)
				return nil, err
			}
		}

		if workflowID != "" {
			if err := s.checkRunFailed(pollCtx, workflowID); err != nil {
				spinner.Fail(err)
				return nil, err
			}
		}

		select {
		case <-pollCtx.Done():
			err := errs.NewUserFacing("timed out after %s waiting for the config to sync", defaultConfigSyncLimit)
			spinner.Fail(err)
			return nil, err
		case <-time.After(defaultConfigSyncPoll):
		}
	}
}

func syncFailureMessage(appConfig *models.AppAppConfig) string {
	if appConfig.StatusDescription != "" {
		return appConfig.StatusDescription
	}
	return fmt.Sprintf("app config %s failed to sync", appConfig.ID)
}

// parseSyncState reads the persisted sync state; missing or unreadable state
// only costs orphan and scheduled-build reporting.
func parseSyncState(stateJSON string) *sync.State {
	if stateJSON == "" {
		return &sync.State{}
	}

	var state sync.State
	if err := json.Unmarshal([]byte(stateJSON), &state); err != nil {
		return &sync.State{}
	}
	return &state
}

func (s *Service) notifySyncResult(result *sync.Result) {
	if result == nil {
		return
	}
	s.notifyOrphanedComponents(result.OrphanedComponents)
	s.notifyOrphanedActions(result.OrphanedActions)
	s.notifyOrphanedRunbooks(result.OrphanedRunbooks)
}

func (s *Service) notifyOrphanedRunbooks(runbooks map[string]string) {
	if len(runbooks) == 0 {
		return
	}

	msg := "Existing runbook(s) are no longer defined in the config:\n"
	for name, id := range runbooks {
		msg += fmt.Sprintf("Runbook: Name=%s | ID=%s\n", name, id)
	}

	ui.PrintLn(msg)
}
