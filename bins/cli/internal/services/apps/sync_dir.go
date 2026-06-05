package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/pkg/config/sync/apisyncer"
	"github.com/nuonco/nuon/pkg/config/validate"
	"github.com/nuonco/nuon/pkg/errs"
)

const (
	defaultSyncTimeout               time.Duration = time.Minute * 20
	defaultSyncSleep                 time.Duration = time.Second * 20
	componentBuildStatusError                      = "error"
	componentBuildStatusPolicyFailed               = "policy_failed"
	componentBuildStatusBuilding                   = "building"
	componentBuildStatusActive                     = "active"
	componentStatusQueued                          = "queued"
)

// SyncOptions controls how the target app is resolved when syncing a directory.
type SyncOptions struct {
	// AppFlag is the value of the explicit --app-id flag (ID or name). Highest precedence.
	AppFlag string
	// DirExplicit is true when the user passed a positional directory argument.
	// When true, the app is resolved from the directory name (preserving legacy
	// behavior) even if there is a currently-selected app.
	DirExplicit bool
	// Create indicates the app should be created if it does not exist.
	Create bool
}

func (s *Service) DeprecatedSyncDir(ctx context.Context, dir string, version string, opts SyncOptions) error {
	deprecatedWarning := config.ErrConfig{
		Description: "nuon apps sync-dir is deprecated, please use nuon apps sync instead",
		Warning:     true,
		Err:         fmt.Errorf("deprecated command nuon sync-dir"),
	}
	ui.PrintError(deprecatedWarning)
	return s.SyncDir(ctx, dir, version, opts)
}

func (s *Service) SyncDir(ctx context.Context, dir string, version string, opts SyncOptions) error {
	return s.syncDir(ctx, dir, version, opts)
}

func (s *Service) SyncDirWithCreate(ctx context.Context, dir string, version string, opts SyncOptions) error {
	opts.Create = true
	return s.syncDir(ctx, dir, version, opts)
}

func (s *Service) syncDir(ctx context.Context, dir string, version string, opts SyncOptions) error {
	ui.PrintLn("syncing directory from " + dir)

	appID, err := s.resolveSyncAppID(ctx, dir, opts)
	if err != nil {
		return ui.PrintError(err)
	}

	cfg, err := parse.ParseDir(ctx, parse.ParseConfig{
		Dirname:       dir,
		V:             validator.New(),
		FileProcessor: func(name string, obj map[string]any) map[string]any { return obj },
	})
	if err != nil {
		return ui.PrintError(err)
	}

	if s.cfg.Debug {
		ui.PrintJSON(cfg)
	}

	ui.PrintLn("validating configs")
	err = validate.Validate(ctx, s.v, cfg)
	if err != nil {
		if config.IsWarningErr(err) {
			ui.PrintError(err)
		} else {
			s.checkSchemaCompatibility(ctx)
			return ui.PrintError(err)
		}
	}
	ui.PrintLn("all configs valid")

	// TODO(onprem): remove this after a few releases
	if len(cfg.Installs) > 0 {
		ui.PrintWarning("deprecated: skipped syncing installs from app config. to sync these installs, switch to 'nuon installs sync' command.")
	}

	for _, runbook := range cfg.Runbooks {
		for _, msg := range runbook.DeprecationWarnings {
			ui.PrintWarning("deprecated: " + msg)
		}
	}

	syncer := apisyncer.New(s.api, appID, version, cfg)
	err = syncer.Sync(ctx)
	if err != nil {
		return ui.PrintError(err)
	}

	if err := s.api.UpdateAppConfigInstalls(ctx, appID, syncer.GetAppConfigID(), &models.ServiceUpdateAppConfigInstallsRequest{
		UpdateAll: true,
	}); err != nil {
		return err
	}

	ui.PrintSuccess("successfully synced " + dir)
	s.notifyOrphanedComponents(syncer.OrphanedComponents())
	s.notifyOrphanedActions(syncer.OrphanedActions())

	cmpsScheduled := syncer.GetComponentsScheduled()
	if len(cmpsScheduled) == 0 {
		return nil
	}

	if err := s.pollComponentBuilds(ctx, cmpsScheduled); err != nil {
		return errors.Wrap(err, "unable to poll builds")
	}

	return nil
}

// resolveSyncAppID determines which app the sync should target.
//
// Precedence:
//  1. Explicit --app-id flag (opts.AppFlag).
//  2. Positional directory argument: app name derived from the directory.
//     This preserves the legacy `nuon apps sync <dir>` behavior.
//  3. Currently-selected app (from `nuon apps select`).
//  4. Fallback: app name derived from the (cwd) directory.
//
// When the dir-name path is taken and a different app is currently selected,
// a warning is printed so the discrepancy is not silent.
func (s *Service) resolveSyncAppID(ctx context.Context, dir string, opts SyncOptions) (string, error) {
	selectedAppID := s.getAppID()

	// (1) --app-id flag wins.
	if opts.AppFlag != "" {
		appID, err := s.resolveOrCreateApp(ctx, opts.AppFlag, opts.Create)
		if err != nil {
			return "", err
		}
		if selectedAppID != "" && selectedAppID != appID {
			ui.PrintWarning(fmt.Sprintf("--app-id %q overrides selected app %s", opts.AppFlag, selectedAppID))
		}
		return appID, nil
	}

	// (2) Positional dir argument: use dir-name (legacy behavior).
	if opts.DirExplicit {
		appID, appName, err := s.resolveFromDirName(ctx, dir, opts.Create)
		if err != nil {
			return "", err
		}
		if selectedAppID != "" && selectedAppID != appID {
			ui.PrintWarning(fmt.Sprintf("selected app is %s but syncing app %q (derived from directory); pass --app-id to override", selectedAppID, appName))
		}
		return appID, nil
	}

	// (3) Selected app from `nuon apps select`.
	if selectedAppID != "" {
		appID, err := lookup.AppID(ctx, s.api, selectedAppID)
		if err != nil {
			return "", errs.WithUserFacing(err, "error looking up selected app")
		}
		return appID, nil
	}

	// (4) Fallback: cwd dir-name (current default behavior).
	appID, _, err := s.resolveFromDirName(ctx, dir, opts.Create)
	return appID, err
}

func (s *Service) resolveFromDirName(ctx context.Context, dir string, create bool) (string, string, error) {
	appName, err := parse.AppNameFromDirName(dir)
	if err != nil {
		return "", "", errs.WithUserFacing(err, "error parsing app name from directory")
	}
	appID, err := s.resolveOrCreateApp(ctx, appName, create)
	if err != nil {
		return "", appName, err
	}
	return appID, appName, nil
}

// resolveOrCreateApp looks up an app by ID or name. If not found and create is
// true, it creates the app using nameOrID as the name and returns the new ID.
func (s *Service) resolveOrCreateApp(ctx context.Context, nameOrID string, create bool) (string, error) {
	appID, err := lookup.AppID(ctx, s.api, nameOrID)
	if err == nil {
		return appID, nil
	}
	if !create {
		return "", errs.WithUserFacing(err, "error looking up app id")
	}

	ui.PrintLn(fmt.Sprintf("app %q not found, creating it", nameOrID))
	if err := s.Create(ctx, nameOrID, false, true); err != nil {
		return "", err
	}

	appID, err = lookup.AppID(ctx, s.api, nameOrID)
	if err != nil {
		return "", errs.WithUserFacing(err, "error looking up app id after creation")
	}
	return appID, nil
}

func (s *Service) notifyOrphanedComponents(cmps map[string]string) {
	if len(cmps) == 0 {
		return
	}

	msg := "Existing component(s) are no longer defined in the config:\n"

	for name, id := range cmps {
		msg += fmt.Sprintf("Component: Name=%s | ID=%s\n", name, id)
	}

	ui.PrintLn(msg)
}

func (s *Service) notifyOrphanedActions(actions map[string]string) {
	if len(actions) == 0 {
		return
	}

	msg := "Existing action(s) are no longer defined in the config:\n"

	for name, id := range actions {
		msg += fmt.Sprintf("Action: Name=%s | ID=%s\n", name, id)
	}

	ui.PrintLn(msg)
}
