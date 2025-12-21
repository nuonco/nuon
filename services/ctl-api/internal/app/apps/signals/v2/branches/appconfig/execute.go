package appconfig

import (
	"encoding/json"
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	l := workflow.GetLogger(ctx)

	run, err := activities.AwaitGetAppBranchRunWithCommitByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run with commit: %w", err)
	}

	commitSHA := run.VCSConnectionCommit.SHA

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

	cloneResult, err := activities.AwaitCloneRepo(ctx, activities.CloneRepoRequest{
		RunID:       s.RunID,
		VcsConfigID: vcsConfigID,
		CommitSHA:   commitSHA,
	})
	if err != nil {
		return fmt.Errorf("unable to clone repo: %w", err)
	}

	sourceDir := cloneResult.SourceDir

	l.Info("repo cloned successfully",
		"app_branch_id", branch.ID,
		"commit_sha", commitSHA,
		"source_dir", sourceDir)

	intermediateConfig, err := activities.AwaitFetchIntermediateConfig(ctx, activities.FetchIntermediateConfigRequest{
		SourceDir: sourceDir,
	})
	if err != nil {
		return fmt.Errorf("unable to fetch intermediate config: %w", err)
	}

	l.Info("intermediate config fetched",
		"app_branch_id", branch.ID,
		"commit_sha", commitSHA,
		"config_version", intermediateConfig.Version,
		"num_components", len(intermediateConfig.Components))

	configJSON, err := json.Marshal(intermediateConfig)
	if err != nil {
		return fmt.Errorf("unable to serialize intermediate config: %w", err)
	}

	createResp, err := activities.AwaitCreateAppConfig(ctx, activities.CreateAppConfigRequest{
		Req: &activities.CreateAppConfigInput{
			AppID:                  branch.AppID,
			OrgID:                  branch.OrgID,
			AppBranchID:            branch.ID,
			CreatedByID:            branch.CreatedByID,
			IntermediateConfigJSON: string(configJSON),
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create app config: %w", err)
	}

	l.Info("app config created",
		"app_config_id", createResp.AppConfigID,
		"app_branch_id", branch.ID)

	syncResp, err := activities.AwaitSyncAppConfig(ctx, activities.SyncAppConfigRequest{
		Req: &activities.SyncAppConfigInput{
			AppConfigID: createResp.AppConfigID,
			AppID:       branch.AppID,
			AppBranchID: branch.ID,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to sync app config: %w", err)
	}

	l.Info("app config synced",
		"app_config_id", syncResp.AppConfigID,
		"component_count", len(syncResp.ComponentIDs))

	if err := activities.AwaitUpdateAppBranchRunAppConfig(ctx, activities.UpdateAppBranchRunAppConfigRequest{
		Req: &activities.UpdateAppBranchRunAppConfigInput{
			RunID:       s.RunID,
			AppConfigID: syncResp.AppConfigID,
		},
	}); err != nil {
		return fmt.Errorf("unable to update run with app config ID: %w", err)
	}

	return nil
}
