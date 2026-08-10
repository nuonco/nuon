package fetchcommit

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	branchactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	installsConfig, err := branchactivities.AwaitGetAppInstallsConfigByAppID(ctx, s.AppID)
	if err != nil {
		return fmt.Errorf("unable to get app installs config: %w", err)
	}

	if !installsConfig.Found {
		s.updateStepMetadata(ctx, map[string]any{"description": "no installs config found"})
		return nil
	}

	if installsConfig.VCSConfigID == "" {
		return fmt.Errorf("no VCS config found for app installs config %s — ensure VCS config is created with the installs config", installsConfig.ID)
	}

	vcsConnectionID := ""
	if installsConfig.VCSConnectionID != nil {
		vcsConnectionID = *installsConfig.VCSConnectionID
	}

	commitResult, err := branchactivities.AwaitFetchInstallSyncCommit(ctx, &branchactivities.FetchInstallSyncCommitInput{
		AppInstallConfigSyncID: s.AppInstallConfigSyncID,
		VCSConfigID:            installsConfig.VCSConfigID,
		CommitSHA:              s.CommitSHA,
		Branch:                 installsConfig.Branch,
	})
	if err != nil {
		return fmt.Errorf("unable to fetch commit: %w", err)
	}

	_, err = branchactivities.AwaitCloneInstallsRepo(ctx, &branchactivities.CloneInstallsRepoInput{
		AppInstallConfigSyncID: s.AppInstallConfigSyncID,
		VCSType:                installsConfig.VCSType,
		VCSConnectionID:        vcsConnectionID,
		Repo:                   installsConfig.Repo,
		Branch:                 installsConfig.Branch,
		CommitSHA:              commitResult.SHA,
	})
	if err != nil {
		return fmt.Errorf("unable to clone installs repo: %w", err)
	}

	short := commitResult.SHA
	if len(short) > 8 {
		short = short[:8]
	}

	meta := map[string]any{
		"repo":         installsConfig.Repo,
		"branch":       installsConfig.Branch,
		"directory":    installsConfig.Directory,
		"triggered_by": s.TriggeredBy,
		"description":  fmt.Sprintf("fetched commit %s", short),
		"commit_sha":   commitResult.SHA,
	}
	if commitResult.Message != "" {
		meta["commit_message"] = commitResult.Message
	}
	if commitResult.Author != "" {
		meta["author_name"] = commitResult.Author
	}

	s.updateStepMetadata(ctx, meta)
	logger.Info("fetched commit and cloned repo",
		"repo", installsConfig.Repo,
		"sha", commitResult.SHA)
	return nil
}

func (s *Signal) updateStepMetadata(ctx workflow.Context, metadata map[string]any) {
	if s.StepID == "" {
		return
	}
	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:   app.StatusInProgress,
			Metadata: metadata,
		},
	})
}
