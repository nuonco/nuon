package planinstallgroup

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/signals/installconfigdiff"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/callback"
	sharedactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	run, err := activities.AwaitGetAppBranchRunByIDByRunID(ctx, s.RunID)
	if err != nil {
		return fmt.Errorf("unable to get app branch run: %w", err)
	}

	installIDs, groupName, err := s.resolveInstallIDs(ctx)
	if err != nil {
		return err
	}

	if len(installIDs) == 0 {
		logger.Info("no installs in group, skipping plan")
		return nil
	}

	s.updatePlanMetadata(ctx, groupName, len(installIDs), nil)

	type pendingDiff struct {
		installID string
		cb        callback.Ref
	}
	pending := make([]pendingDiff, 0, len(installIDs))

	for _, installID := range installIDs {
		cb := callback.New(ctx, installID+"-plan")

		_, err = sharedactivities.AwaitEnqueueSignalToOwner(ctx, &sharedactivities.EnqueueSignalToOwnerRequest{
			OwnerID:   installID,
			OwnerType: "installs",
			QueueName: "install-signals",
			Signal: &installconfigdiff.Signal{
				InstallID:      installID,
				NewAppConfigID: run.AppConfigID,
			},
			Callback: cb,
		})
		if err != nil {
			return fmt.Errorf("install %s: unable to enqueue config diff signal: %w", installID, err)
		}

		pending = append(pending, pendingDiff{installID: installID, cb: cb})
	}

	type installPlanEntry struct {
		InstallID string `json:"install_id"`
		Status    string `json:"status"`
	}

	results := make([]installPlanEntry, 0, len(pending))
	var errs []error

	for _, p := range pending {
		if _, err := callback.AwaitWithTimeout(ctx, p.cb, callback.FallbackAwaitTimeout); err != nil {
			errs = append(errs, fmt.Errorf("install %s: config diff failed: %w", p.installID, err))
			results = append(results, installPlanEntry{InstallID: p.installID, Status: "error"})
		} else {
			results = append(results, installPlanEntry{InstallID: p.installID, Status: "success"})
		}
	}

	s.updatePlanMetadata(ctx, groupName, len(installIDs), results)

	if len(errs) > 0 {
		return fmt.Errorf("plan had %d error(s): %v", len(errs), errs)
	}

	logger.Info("install group plan completed",
		"install_group_id", s.InstallGroupID,
		"install_count", len(installIDs),
	)

	return nil
}

func (s *Signal) resolveInstallIDs(ctx workflow.Context) ([]string, string, error) {
	logger := workflow.GetLogger(ctx)

	group, err := activities.AwaitGetInstallGroupByID(ctx, s.InstallGroupID)
	if err != nil {
		return nil, "", fmt.Errorf("unable to get install group: %w", err)
	}

	if group.LabelSelector == nil {
		logger.Info("planning install group",
			"install_group_id", group.ID,
			"install_group_name", group.Name,
			"install_count", len(group.InstallIDs),
		)
		return group.InstallIDs, group.Name, nil
	}

	branch, err := activities.AwaitGetAppBranchByIDByAppBranchID(ctx, s.AppBranchID)
	if err != nil {
		return nil, "", fmt.Errorf("unable to get app branch for label resolution: %w", err)
	}

	resolved, err := activities.AwaitResolveInstallGroupInstalls(ctx, &activities.ResolveInstallGroupInstallsInput{
		AppID:    branch.AppID,
		GroupID:  group.ID,
		Selector: group.LabelSelector,
	})
	if err != nil {
		return nil, "", fmt.Errorf("unable to resolve install group labels: %w", err)
	}

	logger.Info("planning install group",
		"install_group_id", group.ID,
		"install_group_name", group.Name,
		"install_count", len(resolved.InstallIDs),
		"resolved_via", "label_selector",
	)

	return resolved.InstallIDs, group.Name, nil
}

func (s *Signal) updatePlanMetadata(ctx workflow.Context, groupName string, totalInstalls int, results any) {
	if s.StepID == "" {
		return
	}

	meta := map[string]any{
		"install_group_name": groupName,
		"total_installs":     totalInstalls,
	}
	if results != nil {
		meta["installs"] = results
	}

	desc := fmt.Sprintf("computing plan for %d installs", totalInstalls)
	if results != nil {
		desc = fmt.Sprintf("plan computed for %d installs", totalInstalls)
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusInProgress,
			StatusHumanDescription: desc,
			Metadata:               meta,
		},
	})
}
