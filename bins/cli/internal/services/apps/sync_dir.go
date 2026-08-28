package apps

import (
	"context"
	"fmt"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"

	"github.com/nuonco/nuon/bins/cli/internal/lookup"
	"github.com/nuonco/nuon/bins/cli/internal/ui"
	"github.com/nuonco/nuon/bins/cli/internal/ui/bubbles"
	"github.com/nuonco/nuon/pkg/cli/styles"
	"github.com/nuonco/nuon/pkg/config"
	"github.com/nuonco/nuon/pkg/config/parse"
	"github.com/nuonco/nuon/pkg/config/sync"
	"github.com/nuonco/nuon/pkg/config/validate"
	"github.com/nuonco/nuon/pkg/errs"
)

const (
	defaultSyncTimeout               time.Duration = time.Minute * 20
	defaultSyncSleep                 time.Duration = time.Second * 20
	componentBuildStatusError                      = "error"
	componentBuildStatusPolicyFailed               = "policy_failed"
	componentBuildStatusActive                     = "active"
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
	// AppBranch triggers interactive branch selection when true.
	AppBranch bool
	// Preview creates a plan-only run (no apply). Only used with Branch or AppBranch.
	Preview bool
	// AutoApprove skips the branch run's approval gate before each install group
	// deploys. Without it the gate follows the targeted installs' approval option.
	AutoApprove bool
	// PrintJSON emits a machine-readable result on success (--output json/agent).
	PrintJSON bool
	// NoWait skips waiting for scheduled component builds to complete; the
	// exit code then reflects the sync only.
	NoWait bool
}

// syncResult is the machine-readable summary emitted via ui.PrintJSON when
// SyncOptions.PrintJSON is set, so --output agent gets a success envelope.
type syncResult struct {
	AppID    string            `json:"app_id"`
	Dir      string            `json:"dir"`
	BranchID string            `json:"branch_id,omitempty"`
	RunID    string            `json:"run_id,omitempty"`
	Builds   *syncBuildsResult `json:"builds,omitempty"`
}

// syncBuildsResult summarizes the component builds the sync scheduled.
type syncBuildsResult struct {
	Scheduled  int            `json:"scheduled"`
	Waited     bool           `json:"waited"`
	Components []BuildOutcome `json:"components,omitempty"`
}

// buildsFailedExitCode signals that the config synced but the scheduled
// component builds failed, were policy-blocked, or timed out. Exit 1 remains
// "sync failed" and exit 2 is the read-only guardrail.
const buildsFailedExitCode = 3

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

	s.warnIfCLIOutdated(ctx)

	parseResult, err := parse.ParseDirWithSource(ctx, parse.ParseConfig{
		Dirname:       dir,
		V:             validator.New(),
		FileProcessor: func(name string, obj map[string]any) map[string]any { return obj },
	})
	if err != nil {
		return ui.PrintError(err)
	}
	cfg := parseResult.Config
	var sourceArchive *config.SourceArchive
	if cfg.CustomerManaged != nil {
		sourceArchive = parseResult.Source
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

	var branchID string
	switch {
	case opts.Branch != "":
		var branchErr error
		branchID, branchErr = s.resolveAppBranchID(ctx, appID, opts.Branch)
		if branchErr != nil {
			return ui.PrintError(branchErr)
		}
		ui.PrintLn(fmt.Sprintf("targeting app branch %q", opts.Branch))
	case opts.AppBranch:
		var branchErr error
		branchID, branchErr = s.selectAppBranch(ctx, appID)
		if branchErr != nil {
			return ui.PrintError(branchErr)
		}
	default:
		var branchErr error
		branchID, branchErr = s.resolveDefaultBranchID(ctx, appID)
		if branchErr != nil {
			return ui.PrintError(branchErr)
		}
	}

	appConfig, err := s.createConfig(ctx, appID, version, cfg, sourceArchive, branchID, opts.Preview)
	if err != nil {
		return ui.PrintError(err)
	}

	if branchID != "" {
		result, branchErr := s.syncViaBranchRun(ctx, appID, branchID, dir, appConfig, opts)
		if branchErr != nil {
			return ui.PrintError(branchRunSyncErr(branchErr, result))
		}
		if opts.PrintJSON {
			ui.PrintJSON(*result)
		}
		return nil
	}

	state, err := s.syncConfig(ctx, appID, appConfig, opts)
	if err != nil {
		return ui.PrintError(err)
	}

	ui.PrintSuccess("successfully synced " + dir)
	s.notifySyncResult(state.Result)

	result := syncResult{AppID: appID, Dir: dir}
	var cmpsScheduled []sync.ComponentState
	if state.Result != nil {
		cmpsScheduled = state.Result.ComponentsScheduled
	}
	if len(cmpsScheduled) == 0 {
		if opts.PrintJSON {
			ui.PrintJSON(result)
		}
		return nil
	}

	result.Builds = &syncBuildsResult{
		Scheduled: len(cmpsScheduled),
		Waited:    !opts.NoWait,
	}

	if opts.NoWait {
		ui.PrintLn(fmt.Sprintf("%d component build(s) scheduled; not waiting for completion (--no-wait)", len(cmpsScheduled)))
		if opts.PrintJSON {
			ui.PrintJSON(result)
		}
		return nil
	}

	ui.PrintLn(fmt.Sprintf("waiting for %d component build(s) to complete (timeout %s; use --no-wait to skip)", len(cmpsScheduled), defaultSyncTimeout))
	outcomes, pollErr := s.pollComponentBuilds(ctx, cmpsScheduled)
	result.Builds.Components = outcomes
	if pollErr != nil {
		return ui.PrintError(&ui.ErrExitCode{
			Err:  errors.Wrap(pollErr, "app config synced, but component builds did not succeed"),
			Code: "builds_failed",
			Exit: buildsFailedExitCode,
		})
	}

	ui.PrintSuccess("all component builds completed")
	if opts.PrintJSON {
		ui.PrintJSON(result)
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

func (s *Service) selectAppBranch(ctx context.Context, appID string) (string, error) {
	branches, err := s.api.GetAppBranches(ctx, appID)
	if err != nil {
		return "", fmt.Errorf("unable to list app branches: %w", err)
	}

	if len(branches) == 0 {
		return "", fmt.Errorf("no app branches found for this app")
	}

	opts := make([]bubbles.BranchOption, 0, len(branches))
	for _, b := range branches {
		opts = append(opts, bubbles.BranchOption{ID: b.ID, Name: b.Name})
	}

	selected, err := bubbles.SelectBranch(opts, s.cfg.Interactive)
	if err != nil {
		return "", err
	}

	return selected, nil
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
