package installcreate

import (
	"context"
	"fmt"
	"time"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/sdks/nuon-go"
	"github.com/nuonco/nuon/sdks/nuon-go/models"
)

var (
	BranchConfigPoll    = 10 * time.Second
	BranchConfigTimeout = 15 * time.Minute
)

const branchRunsPageSize = 50

// AwaitBranchConfig waits until the branch has an active app config.
// A nil branch skips the wait. Preview runs and finished runs without a config
// are skipped in favor of the newest earlier run that has one; with none, the
// call returns immediately. An in-flight run is polled until its config is active.
// Dry run fails immediately when the config is not already ready.
func AwaitBranchConfig(
	ctx context.Context,
	api nuon.Client,
	appID string,
	branch *models.AppAppBranch,
	dryRun bool,
	log func(string),
) (*models.AppAppBranch, error) {
	if branch == nil {
		return nil, nil
	}

	deadline := time.Now().Add(BranchConfigTimeout)
	current := branch
	for {
		ready, err := branchConfigReady(ctx, api, appID, current)
		if err != nil {
			return nil, err
		}
		if ready != nil {
			return ready, nil
		}
		if dryRun || time.Now().After(deadline) {
			return nil, fmt.Errorf("app branch %q has no app config to apply", current.Name)
		}
		if log != nil {
			log(fmt.Sprintf("waiting for app branch %s config", current.Name))
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(BranchConfigPoll):
		}

		current, err = api.GetAppBranch(ctx, appID, branch.ID)
		if err != nil {
			return nil, fmt.Errorf("error getting app branch %q: %w", branch.Name, err)
		}
		if current == nil {
			return nil, fmt.Errorf("app branch %q was not found", branch.Name)
		}
	}
}

// branchConfigReady reports whether the branch has an active app config. The
// returned branch is non-nil when ready; its LatestRun is the run the config
// belongs to, which differs from the branch's own latest run when that run is
// a preview or finished without producing a config.
func branchConfigReady(ctx context.Context, api nuon.Client, appID string, branch *models.AppAppBranch) (*models.AppAppBranch, error) {
	run, err := branchConfigRun(ctx, api, appID, branch)
	if err != nil || run == nil {
		return nil, err
	}

	cfg, err := api.GetAppConfig(ctx, appID, run.AppConfigID, generics.ToPtr(true))
	if err != nil {
		return nil, fmt.Errorf("error getting app config for branch %q: %w", branch.Name, err)
	}
	if cfg == nil {
		return nil, nil
	}
	switch cfg.Status {
	case models.AppAppConfigStatusActive:
		if run == branch.LatestRun {
			return branch, nil
		}
		resolved := *branch
		resolved.LatestRun = run
		return &resolved, nil
	case models.AppAppConfigStatusError:
		return nil, fmt.Errorf("app branch %q config %s failed to build", branch.Name, cfg.ID)
	default:
		return nil, nil
	}
}

// branchConfigRun returns the run whose app config the branch currently
// serves, or nil when the caller should keep waiting. Preview runs and runs
// that finished without a config (such as ones skipped by ignore_changes_regex)
// never produce the branch's config, so they fall back to the newest earlier
// run that has one instead of being waited on.
func branchConfigRun(ctx context.Context, api nuon.Client, appID string, branch *models.AppAppBranch) (*models.AppAppBranchRun, error) {
	run := branch.LatestRun
	if run == nil {
		return nil, nil
	}
	if hasBranchConfig(run) {
		return run, nil
	}
	preview := isPreviewRun(run)
	if !preview && !branchRunTerminal(run.Status) {
		return nil, nil
	}

	fallback, err := latestConfiguredRun(ctx, api, appID, branch)
	if err != nil {
		return nil, err
	}
	if fallback != nil {
		return fallback, nil
	}
	if !branchRunTerminal(run.Status) {
		return nil, nil
	}
	if branchRunFailed(run.Status) {
		return nil, fmt.Errorf("app branch %q run failed", branch.Name)
	}
	return nil, fmt.Errorf("app branch %q has no app config to apply", branch.Name)
}

func latestConfiguredRun(ctx context.Context, api nuon.Client, appID string, branch *models.AppAppBranch) (*models.AppAppBranchRun, error) {
	workflows, _, err := api.GetAppBranchRunsWithQuery(ctx, appID, branch.ID, &nuon.GetAppBranchRunsQuery{
		Planonly: generics.ToPtr(false),
		Limit:    branchRunsPageSize,
	})
	if err != nil {
		return nil, fmt.Errorf("error listing runs for app branch %q: %w", branch.Name, err)
	}
	for _, workflow := range workflows {
		if workflow == nil {
			continue
		}
		for _, run := range workflow.AppBranchRuns {
			if run != nil && hasBranchConfig(run) {
				return run, nil
			}
		}
	}
	return nil, nil
}

func hasBranchConfig(run *models.AppAppBranchRun) bool {
	return run.AppConfigID != "" && !isPreviewRun(run)
}

func isPreviewRun(run *models.AppAppBranchRun) bool {
	return run.Preview != nil || run.PlanOnly || run.RunType == models.AppAppBranchRunTypeGitDashPreviewDashRun
}

func branchRunTerminal(status string) bool {
	return branchRunFailed(status) || status == "success" || status == "not-attempted"
}

func branchRunFailed(status string) bool {
	switch status {
	case "failed", "error", "cancelled":
		return true
	default:
		return false
	}
}
