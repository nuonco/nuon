package installconfigsync

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	commitResult, err := activities.AwaitFetchLatestInstallsCommitByAppID(ctx, install.AppID)
	if err != nil {
		logger.Warn("unable to fetch latest commit for installs config", "error", err)
	}

	vcsResult, err := activities.AwaitGetInstallsVCSConfigByAppID(ctx, install.AppID)
	if err != nil {
		return fmt.Errorf("unable to get installs VCS config: %w", err)
	}

	if !vcsResult.HasVCSConfig {
		logger.Info("no installs VCS config in app config, skipping")
		s.updateStepStatus(ctx, app.StatusSuccess, "no installs config", nil)
		return nil
	}

	installsDir := vcsResult.InstallsDir
	if installsDir == "" {
		installsDir = "."
	}

	syncInput := &activities.CreateInstallConfigSyncInput{
		InstallID:              s.InstallID,
		AppInstallConfigSyncID: s.AppInstallConfigSyncID,
		AppBranchID:            s.AppBranchID,
		AppBranchConfigID:      s.AppBranchConfigID,
		AppBranchRunID:         s.AppBranchRunID,
		TriggeredBy:            s.TriggeredBy,
	}

	if commitResult != nil && commitResult.CommitID != "" {
		syncInput.VCSConnectionCommitID = commitResult.CommitID
	}

	syncRecord, err := activities.AwaitCreateInstallConfigSync(ctx, syncInput)
	if err != nil {
		return fmt.Errorf("unable to create install config sync record: %w", err)
	}

	// TODO(jm): remove hardcoded local path once network clone is working
	sourceDir := "/Users/jonmorehouse/nuon/kitchen-sink"

	installConfigs, err := activities.AwaitParseInstallConfigs(ctx, activities.ParseInstallConfigsRequest{
		SourceDir: sourceDir,
		Req: &activities.ParseInstallConfigsInput{
			SourceDir:         sourceDir,
			InstallsDirectory: installsDir,
			InstallName:       install.Name,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to parse install configs: %w", err)
	}

	if len(installConfigs.Installs) == 0 {
		logger.Info("no install config found to sync")
		_ = activities.AwaitUpdateInstallConfigSyncStatus(ctx, &activities.UpdateInstallConfigSyncStatusInput{
			InstallConfigSyncID: syncRecord.ID,
			Status:              string(app.StatusSuccess),
			StatusDescription:   "no install config found",
		})
		s.updateStepStatus(ctx, app.StatusSuccess, "no install config found", nil)
		return nil
	}

	synced := 0
	failed := 0
	created := 0

	for _, ic := range installConfigs.Installs {
		result, err := activities.AwaitSyncInstallConfig(ctx, &activities.SyncInstallConfigInput{
			AppID:               install.AppID,
			InstallConfig:       ic.Config,
			InstallConfigSyncID: syncRecord.ID,
			FilePath:            ic.FilePath,
		})
		if err != nil {
			logger.Warn("failed to sync install config",
				"install_name", ic.Config.Name,
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
		Status:              string(finalStatus),
		StatusDescription:   desc,
	})

	s.updateStepStatus(ctx, finalStatus, desc, map[string]any{
		"synced":  synced,
		"created": created,
		"failed":  failed,
		"total":   len(installConfigs.Installs),
	})

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
