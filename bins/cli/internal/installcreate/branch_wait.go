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

// AwaitBranchConfig waits until the branch has an active app config.
// A nil branch skips the wait. A branch whose only run has failed returns
// immediately. An in-flight run is polled until its config is active.
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
		if ready {
			return current, nil
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

func branchConfigReady(ctx context.Context, api nuon.Client, appID string, branch *models.AppAppBranch) (bool, error) {
	run := branch.LatestRun
	if run != nil && branchRunFailed(run.Status) && run.AppConfigID == "" {
		runs, err := api.GetAppBranchRuns(ctx, appID, branch.ID)
		if err != nil {
			return false, fmt.Errorf("error listing runs for app branch %q: %w", branch.Name, err)
		}
		if len(runs) <= 1 {
			return false, fmt.Errorf("app branch %q run failed", branch.Name)
		}
	}
	if run == nil || run.AppConfigID == "" {
		return false, nil
	}
	cfg, err := api.GetAppConfig(ctx, appID, run.AppConfigID, generics.ToPtr(true))
	if err != nil {
		return false, fmt.Errorf("error getting app config for branch %q: %w", branch.Name, err)
	}
	if cfg == nil {
		return false, nil
	}
	switch cfg.Status {
	case models.AppAppConfigStatusActive:
		return true, nil
	case models.AppAppConfigStatusError:
		return false, fmt.Errorf("app branch %q config %s failed to build", branch.Name, cfg.ID)
	default:
		return false, nil
	}
}

func branchRunFailed(status string) bool {
	switch status {
	case "failed", "error", "cancelled":
		return true
	default:
		return false
	}
}
