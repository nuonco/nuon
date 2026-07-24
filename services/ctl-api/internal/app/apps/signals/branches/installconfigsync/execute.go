package installconfigsync

import (
	"fmt"
	"path/filepath"
	"strings"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return fmt.Errorf("unable to get app branch: %w", err)
	}

	vcsResult, err := activities.AwaitGetInstallsVCSConfigByAppID(ctx, branch.AppID)
	if err != nil {
		return fmt.Errorf("unable to get installs VCS config: %w", err)
	}

	if !vcsResult.HasVCSConfig {
		logger.Info("no installs VCS config in app config, skipping install config sync")
		return nil
	}

	installsDir := vcsResult.InstallsDir
	if installsDir == "" {
		installsDir = "."
	}

	runID := s.AppBranchRunID
	if runID == "" {
		runID = s.AppBranchID
	}

	cloneResult, err := activities.LocalAwaitCloneRepo(ctx, activities.CloneRepoRequest{
		RunID:       runID,
		VcsConfigID: vcsResult.VCSConfigID,
		CommitSHA:   s.CommitSHA,
	})
	if err != nil {
		return fmt.Errorf("unable to clone repo: %w", err)
	}

	sourceDir := cloneResult.SourceDir

	installConfigs, err := activities.AwaitParseInstallConfigs(ctx, activities.ParseInstallConfigsRequest{
		SourceDir: sourceDir,
		Req: &activities.ParseInstallConfigsInput{
			SourceDir:         sourceDir,
			InstallsDirectory: installsDir,
			InstallName:       s.InstallName,
			ChangedFiles:      s.ChangedFiles,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to parse install configs: %w", err)
	}

	if len(installConfigs.Installs) == 0 {
		logger.Info("no install configs found to sync")
		s.updateStepStatus(ctx, app.StatusSuccess, "no install configs found", nil)
		return nil
	}

	syncRecord, err := activities.AwaitCreateInstallConfigSync(ctx, &activities.CreateInstallConfigSyncInput{
		AppBranchID:       s.AppBranchID,
		AppBranchConfigID: s.AppBranchConfigID,
		AppBranchRunID:    s.AppBranchRunID,
		CommitSHA:         s.CommitSHA,
		TriggeredBy:       s.TriggeredBy,
		TotalInstalls:     len(installConfigs.Installs),
	})
	if err != nil {
		return fmt.Errorf("unable to create install config sync record: %w", err)
	}

	synced := 0
	failed := 0
	created := 0

	for _, ic := range installConfigs.Installs {
		result, err := activities.AwaitSyncInstallConfig(ctx, &activities.SyncInstallConfigInput{
			AppID:               branch.AppID,
			InstallConfig:       ic.Config,
			InstallConfigSyncID: syncRecord.ID,
			FilePath:            ic.FilePath,
		})
		if err != nil {
			logger.Warn("failed to sync install config",
				"install_name", ic.Config.Name,
				"file_path", ic.FilePath,
				"error", err,
			)
			failed++
			continue
		}

		if result.Changed {
			synced++
			if result.Created {
				created++
			}
		}
	}

	finalStatus := app.StatusSuccess
	desc := fmt.Sprintf("%d installs synced", synced)
	if created > 0 {
		desc += fmt.Sprintf(" (%d created)", created)
	}
	if failed > 0 {
		finalStatus = app.StatusError
		desc += fmt.Sprintf(", %d failed", failed)
	}

	_ = activities.AwaitUpdateInstallConfigSyncStatus(ctx, &activities.UpdateInstallConfigSyncStatusInput{
		InstallConfigSyncID: syncRecord.ID,
		SyncedInstalls:      synced,
		FailedInstalls:      failed,
		Status:              string(finalStatus),
		StatusDescription:   desc,
	})

	s.updateStepStatus(ctx, finalStatus, desc, map[string]any{
		"synced":  synced,
		"created": created,
		"failed":  failed,
		"total":   len(installConfigs.Installs),
	})

	logger.Info("install config sync completed",
		"synced", synced,
		"created", created,
		"failed", failed,
		"total", len(installConfigs.Installs),
	)

	if failed > 0 {
		return fmt.Errorf("install config sync had %d failures", failed)
	}

	return nil
}

func (s *Signal) updateStepStatus(ctx workflow.Context, status app.Status, desc string, metadata map[string]any) {
	if s.StepID == "" {
		return
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 status,
			StatusHumanDescription: desc,
			Metadata:               metadata,
		},
	})
}

func matchesInstallsDirectory(filePath, installsDir string) bool {
	dir := filepath.Dir(filePath)
	installsDir = strings.TrimSuffix(installsDir, "/")
	return dir == installsDir || strings.HasPrefix(dir, installsDir+"/")
}
