package updateappconfig

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-update-app-config"

type Signal struct {
	AppBranchID string `json:"app_branch_id"`
	CommitSHA   string `json:"commit_sha"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.AppBranchID == "" {
		return errors.New("app_branch_id is required")
	}
	if s.CommitSHA == "" {
		return errors.New("commit_sha is required")
	}

	_, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return errors.Wrap(err, "app branch not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	// Get app branch with latest config
	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	// Check if branch has a config with VCS settings
	if len(branch.Configs) == 0 || branch.Configs[0].ConnectedGithubVCSConfig == nil {
		return fmt.Errorf("app branch has no VCS config")
	}

	// Parse nuon.yaml from repo at commit using the config's VCS config ID
	config, err := activities.AwaitParseNuonYamlFromRepo(ctx, activities.ParseNuonYamlFromRepoRequest{
		VcsConfigID: branch.Configs[0].ConnectedGithubVCSConfig.ID,
		CommitSHA:   s.CommitSHA,
	})
	if err != nil {
		return fmt.Errorf("unable to parse nuon.yaml: %w", err)
	}

	// Create new AppConfig
	appConfig, err := activities.AwaitCreateAppConfig(ctx, activities.CreateAppConfigRequest{
		AppID:  branch.AppID,
		Config: config,
	})
	if err != nil {
		return fmt.Errorf("unable to create app config: %w", err)
	}

	// Update app branch last synced commit
	if err := activities.AwaitUpdateAppBranchLastSyncedCommit(ctx, activities.UpdateAppBranchLastSyncedCommitRequest{
		AppBranchID: branch.ID,
		CommitSHA:   s.CommitSHA,
	}); err != nil {
		return fmt.Errorf("unable to update last synced commit: %w", err)
	}

	workflow.GetLogger(ctx).Info("app config updated",
		"app_branch_id", branch.ID,
		"commit_sha", s.CommitSHA,
		"app_config_id", appConfig.ID)

	// TODO: Enqueue build-components signal
	// This will be implemented when build-components signal is ready

	return nil
}
