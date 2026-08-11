package dispatchsyncs

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	branchactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	installhelpers "github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/installconfigsync"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
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

	vcsConnectionID := ""
	if installsConfig.VCSConnectionID != nil {
		vcsConnectionID = *installsConfig.VCSConnectionID
	}

	cloneResult, err := branchactivities.AwaitCloneInstallsRepo(ctx, &branchactivities.CloneInstallsRepoInput{
		AppInstallConfigSyncID: s.AppInstallConfigSyncID,
		VCSType:                installsConfig.VCSType,
		VCSConnectionID:        vcsConnectionID,
		Repo:                   installsConfig.Repo,
		Branch:                 installsConfig.Branch,
		CommitSHA:              s.CommitSHA,
	})
	if err != nil {
		return fmt.Errorf("unable to clone installs repo: %w", err)
	}

	installsDir := s.InstallsDirectory
	if installsDir == "" {
		installsDir = "."
	}

	parsedConfigs, err := branchactivities.AwaitParseInstallConfigs(ctx, branchactivities.ParseInstallConfigsRequest{
		SourceDir: cloneResult.SourceDir,
		Req: &branchactivities.ParseInstallConfigsInput{
			SourceDir:         cloneResult.SourceDir,
			InstallsDirectory: installsDir,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to parse install configs: %w", err)
	}

	installs, err := branchactivities.AwaitGetInstallsForAppByAppID(ctx, s.AppID)
	if err != nil {
		return fmt.Errorf("unable to get installs for app: %w", err)
	}

	existingByName := make(map[string]app.Install, len(installs))
	for _, inst := range installs {
		existingByName[inst.Name] = inst
	}

	enqueued := 0
	for _, pc := range parsedConfigs.Installs {
		if pc.Config == nil {
			continue
		}
		inst, ok := existingByName[pc.Config.Name]
		if !ok {
			continue
		}

		_, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   inst.ID,
			OwnerType: "installs",
			QueueName: installhelpers.InstallSignalsQueueName,
			Signal: &installconfigsync.Signal{
				InstallID:              inst.ID,
				AppInstallConfigSyncID: s.AppInstallConfigSyncID,
				CommitSHA:              s.CommitSHA,
				TriggeredBy:            s.TriggeredBy,
				SourceDir:              cloneResult.SourceDir,
			},
		})
		if err != nil {
			logger.Warn("failed to enqueue install-config-sync",
				"install_id", inst.ID,
				"install_name", inst.Name,
				"error", err,
			)
			continue
		}
		enqueued++
	}

	desc := fmt.Sprintf("dispatched sync to %d installs", enqueued)
	if enqueued == 0 && len(parsedConfigs.Installs) > 0 {
		desc = "no installs to sync"
	}
	if enqueued == 0 && len(parsedConfigs.Installs) == 0 {
		desc = "no install configs found"
	}

	s.updateStepMetadata(ctx, map[string]any{
		"enqueued":    enqueued,
		"total":       len(parsedConfigs.Installs),
		"description": desc,
	})

	logger.Info(desc)
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
