package pininstallgroup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/installgroups"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/updateappconfig"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/queuenames"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

type pinnedInstall struct {
	installID string
	cb        callback.Ref
}

// installVersionStatusVersion gates the per-install config-version status
// update below; in-flight histories never scheduled this activity and must
// keep replaying without it.
const installVersionStatusVersion = "pin-install-app-config-version-status-v1"

// updateInstallAppConfigVersionStatus moves the pinned InstallAppConfigVersion
// row created at the start of pinning to a terminal status. Without this it
// stays "pending" forever once the pin signal completes.
func (s *Signal) updateInstallAppConfigVersionStatus(ctx workflow.Context, installID string, status app.Status, desc string) {
	if workflow.GetVersion(ctx, installVersionStatusVersion, workflow.DefaultVersion, 1) == workflow.DefaultVersion {
		return
	}
	_, _ = activities.AwaitUpdateInstallAppConfigVersionStatus(ctx, &activities.UpdateInstallAppConfigVersionStatusInput{
		AppBranchRunID: s.RunID,
		InstallID:      installID,
		Status:         status,
		StatusDesc:     desc,
	})
}

func (s *Signal) resolveInstallIDs(ctx workflow.Context) ([]string, string, error) {
	resolved, err := installgroups.Resolve(ctx, s.InstallGroupID, s.AppBranchID)
	if err != nil {
		return nil, "", err
	}
	return resolved.InstallIDs, resolved.GroupName, nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	if run.NoConfigChanges && !run.Force {
		logger.Info("no config changes, skipping pin install group")
		return nil
	}

	if run.AppConfigID == "" {
		return fmt.Errorf("app branch run %s has no app config ID", s.RunID)
	}

	installIDs, groupName, err := s.resolveInstallIDs(ctx)
	if err != nil {
		return err
	}

	if len(installIDs) == 0 {
		logger.Info("no installs in group, skipping pin")
		return nil
	}

	groupRunResult, err := activities.AwaitCreateInstallGroupRun(ctx, &activities.CreateInstallGroupRunInput{
		AppBranchRunID:   s.RunID,
		InstallGroupID:   s.InstallGroupID,
		InstallGroupName: groupName,
		TotalInstalls:    len(installIDs),
	})
	if err != nil {
		return fmt.Errorf("unable to create install group run: %w", err)
	}
	groupRunID := groupRunResult.InstallGroupRunID

	installEntries := make([]app.InstallGroupRunInstall, len(installIDs))
	for i, id := range installIDs {
		installEntries[i] = app.InstallGroupRunInstall{
			InstallID: id,
			Status:    string(app.StatusInProgress),
			Phase:     app.InstallGroupRunPhasePin,
		}
	}

	s.writePinMetadata(ctx, groupName, installEntries)

	pending := make([]pinnedInstall, 0, len(installIDs))
	for _, installID := range installIDs {
		if _, err := activities.AwaitCreateInstallAppConfigVersion(ctx, &activities.CreateInstallAppConfigVersionInput{
			InstallID:      installID,
			NewAppConfigID: run.AppConfigID,
			AppBranchRunID: s.RunID,
			InstallGroupID: s.InstallGroupID,
		}); err != nil {
			return fmt.Errorf("install %s: unable to create pinned app config version: %w", installID, err)
		}

		cb := callback.New(ctx, installID)
		if _, err := sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   installID,
			OwnerType: "installs",
			QueueName: queuenames.InstallSignalsQueueName,
			Callback:  cb,
			Signal: &updateappconfig.Signal{
				InstallID:      installID,
				NewAppConfigID: run.AppConfigID,
				AppBranchRunID: s.RunID,
				InstallGroupID: s.InstallGroupID,
				TriggeredBy:    "app-branch",
				Metadata: map[string]string{
					"source":        "app-branch",
					"do_not_deploy": "true",
				},
			},
		}); err != nil {
			return fmt.Errorf("install %s: unable to enqueue update-app-config signal: %w", installID, err)
		}

		pending = append(pending, pinnedInstall{installID: installID, cb: cb})
	}

	var errCount int
	for i, p := range pending {
		_, err := callback.AwaitWithTimeout(ctx, p.cb, callback.FallbackAwaitTimeout)
		if err != nil {
			logger.Warn("pin install failed", "install_id", p.installID, "error", err)
			installEntries[i].Status = string(app.StatusError)
			s.updateInstallAppConfigVersionStatus(ctx, p.installID, app.StatusError, err.Error())
			errCount++
			continue
		}
		installEntries[i].Status = string(app.StatusPinned)
		s.updateInstallAppConfigVersionStatus(ctx, p.installID, app.StatusSuccess, "app version pinned")
	}

	now := workflow.Now(ctx)
	completed := len(installEntries) - errCount
	desc := fmt.Sprintf("pinned app version on %d/%d installs", completed, len(installEntries))
	finalStatus := app.StatusSuccess
	if errCount > 0 {
		finalStatus = app.StatusError
		desc += fmt.Sprintf(" (%d failed)", errCount)
	}

	_ = activities.AwaitUpdateInstallGroupRun(ctx, &activities.UpdateInstallGroupRunInput{
		InstallGroupRunID: groupRunID,
		Installs:          installEntries,
		CompletedInstalls: completed,
		FailedInstalls:    errCount,
		CompletedAt:       &now,
		Status: app.CompositeStatus{
			Status:                 finalStatus,
			StatusHumanDescription: desc,
		},
	})

	s.writePinMetadata(ctx, groupName, installEntries)

	if errCount > 0 {
		return fmt.Errorf("pin install group had %d errors out of %d installs", errCount, len(installEntries))
	}

	return nil
}

func (s *Signal) writePinMetadata(ctx workflow.Context, groupName string, installEntries []app.InstallGroupRunInstall) {
	if s.StepID == "" {
		return
	}

	installs := make([]any, 0, len(installEntries))
	completed := 0
	failed := 0
	for _, e := range installEntries {
		if e.Status == string(app.StatusPinned) {
			completed++
		}
		if e.Status == string(app.StatusError) {
			failed++
		}
		installs = append(installs, map[string]any{
			"install_id": e.InstallID,
			"status":     e.Status,
		})
	}

	desc := fmt.Sprintf("pinning app version on %d installs", len(installEntries))
	if completed > 0 || failed > 0 {
		desc = fmt.Sprintf("pinned app version on %d/%d installs", completed, len(installEntries))
		if failed > 0 {
			desc += fmt.Sprintf(" (%d failed)", failed)
		}
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: desc,
			Metadata: map[string]any{
				"install_group_name": groupName,
				"total_installs":     len(installEntries),
				"installs":           installs,
			},
		},
	})
}
