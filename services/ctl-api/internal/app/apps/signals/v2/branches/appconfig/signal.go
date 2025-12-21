package appconfig

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
)

const SignalType signal.SignalType = "app-branch-app-config"

type Signal struct {
	AppBranchID string `json:"app_branch_id" validate:"required"`
	CommitSHA   string `json:"commit_sha" validate:"required"`
}

var _ signal.Signal = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) Validate(ctx workflow.Context) error {
	v := validator.New()
	if err := v.Struct(s); err != nil {
		return errors.Wrap(err, "validation failed")
	}

	_, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return errors.Wrap(err, "app branch not found")
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	if len(branch.Configs) == 0 {
		return fmt.Errorf("app branch has no config")
	}

	var vcsConfigID string
	if cfg := branch.Configs[0].ConnectedGithubVCSConfig; cfg != nil {
		vcsConfigID = cfg.ID
	} else if cfg := branch.Configs[0].PublicGitVCSConfig; cfg != nil {
		vcsConfigID = cfg.ID
	} else {
		return fmt.Errorf("app branch has no VCS config")
	}

	// Fetch the intermediate config from the repo
	appConfig, err := activities.AwaitFetchIntermediateConfig(ctx, activities.FetchIntermediateConfigRequest{
		VcsConfigID: vcsConfigID,
		CommitSHA:   s.CommitSHA,
	})
	if err != nil {
		return fmt.Errorf("unable to fetch intermediate config: %w", err)
	}

	// Update app branch last synced commit
	if err := activities.AwaitUpdateAppBranchLastSyncedCommit(ctx, activities.UpdateAppBranchLastSyncedCommitRequest{
		AppBranchID: branch.ID,
		CommitSHA:   s.CommitSHA,
	}); err != nil {
		return fmt.Errorf("unable to update last synced commit: %w", err)
	}

	workflow.GetLogger(ctx).Info("intermediate config fetched",
		"app_branch_id", branch.ID,
		"commit_sha", s.CommitSHA,
		"config_version", appConfig.Version,
		"num_components", len(appConfig.Components))

	// TODO: sync the intermediate config via separate activities as a follow-up

	return nil
}
