package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/sdks/nuon-go/models"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/pkg/cli/styles"
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
	// AppFlag is the resolved value of the --app-id flag (ID or name) or picked from context.
	AppFlag string
	// Force, when true, suppresses the directory-mismatch confirmation prompt
	// and syncs to AppFlag regardless of the working directory name.
	Force bool
	// Create indicates the app should be created if it does not exist.
	Create bool
	// Branch optionally targets a specific app branch for this sync.
	Branch string
	// Preview creates a plan-only run (no apply). Only used with Branch.
	Preview bool
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

	var syncerOpts []apisyncer.SyncerOption
	if opts.Branch != "" {
		branchID, branchErr := s.resolveAppBranchID(ctx, appID, opts.Branch)
		if branchErr != nil {
			return ui.PrintError(branchErr)
		}
		syncerOpts = append(syncerOpts, apisyncer.WithAppBranch(branchID, opts.Preview))
		ui.PrintLn(fmt.Sprintf("targeting app branch %q", opts.Branch))
	}

	syncer := apisyncer.New(s.api, appID, version, cfg, syncerOpts...)
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
// Algorithm:
//  1. AppFlag empty (no selection, no explicit --app-id): derive from the
//     working directory name (legacy default).
//  2. AppFlag set (auto-bound from selected app OR explicit --app-id):
//     resolve it and check that the directory name resolves to the same app.
//     - On match: proceed.
//     - On mismatch + --force: warn and proceed.
//     - On mismatch + interactive: prompt for confirmation.
//     - On mismatch + non-interactive: error, suggest --force.
func (s *Service) resolveSyncAppID(ctx context.Context, dir string, opts SyncOptions) (string, error) {
	// (1) No app-id context at all → legacy dir-name behavior.
	if opts.AppFlag == "" {
		appID, _, err := s.resolveFromDirName(ctx, dir, opts.Create)
		return appID, err
	}

	// (2) App-id is set; resolve it to a concrete app ID.
	targetAppID, err := s.resolveOrCreateApp(ctx, opts.AppFlag, opts.Create)
	if err != nil {
		return "", err
	}

	// Compare against the directory-derived app.
	appName, err := parse.AppNameFromDirName(dir)
	if err != nil {
		return "", errs.WithUserFacing(err, "error parsing app name from directory")
	}
	dirAppID, dirErr := lookup.AppID(ctx, s.api, appName)
	if dirErr == nil && dirAppID == targetAppID {
		return targetAppID, nil // match
	}

	// Mismatch path. Fetch the target app's name so messages are friendly
	// even when AppFlag is an opaque ID (e.g. auto-bound from ~/.nuon).
	targetLabel := opts.AppFlag
	if app, err := s.api.GetApp(ctx, targetAppID); err == nil && app != nil && app.Name != "" {
		targetLabel = app.Name
	}
	notice := fmt.Sprintf("directory %q does not match the selected app %q", appName, targetLabel)

	if opts.Force {
		ui.PrintWarning(notice + "; --force in effect, syncing to selected app")
		return targetAppID, nil
	}

	if !s.cfg.Interactive {
		return "", errs.NewUserFacing(
			"%s; pass --force to sync to selected app, or pass matching app ID or name with --app-id",
			notice,
		)
	}

	fmt.Println(styles.TextDim.Render("  " + notice))
	confirmed, err := bubbles.InlineConfirm(
		fmt.Sprintf("Sync directory named %q to app %q?", appName, targetLabel),
		false,
		s.cfg.Interactive,
	)
	if err != nil {
		return "", err
	}
	if !confirmed {
		return "", errs.NewUserFacing("sync cancelled")
	}
	return targetAppID, nil
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

// resolveAppBranchID resolves a branch name or ID to a branch ID.
func (s *Service) resolveAppBranchID(ctx context.Context, appID, branchNameOrID string) (string, error) {
	branches, err := s.api.GetAppBranches(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("unable to list app branches: %w", err)
	}

	for _, b := range branches {
		if b.ID == branchNameOrID || b.Name == branchNameOrID {
			return b.ID, nil
		}
	}

	return "", fmt.Errorf("app branch %q not found", branchNameOrID)
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
