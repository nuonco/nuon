package comparison

import (
	"fmt"

	"go.temporal.io/sdk/workflow"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/activities"
	statusactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/status/activities"
)

func (s *Signal) Execute(ctx workflow.Context) error {
	logger := workflow.GetLogger(ctx)

	out, err := activities.AwaitComputeAndStoreAppBranchRunComparison(ctx, &activities.ComputeAndStoreAppBranchRunComparisonInput{
		AppBranchID: s.AppBranchID,
		RunID:       s.RunID,
	})
	if err != nil {
		return fmt.Errorf("unable to compute run comparison: %w", err)
	}

	if s.StepID == "" {
		return nil
	}

	meta := map[string]any{
		"head_run_id": s.RunID,
	}
	desc := "differences computed"
	if out != nil {
		if out.Skipped {
			desc = "differences skipped"
			if out.SkipReason != "" {
				meta["skip_reason"] = out.SkipReason
				desc = fmt.Sprintf("differences skipped: %s", out.SkipReason)
			}
		}
		if out.BaseRunID != "" {
			meta["base_run_id"] = out.BaseRunID
		}
		if out.BaseSHA != "" {
			meta["base_sha"] = out.BaseSHA
		}
		if out.HeadSHA != "" {
			meta["head_sha"] = out.HeadSHA
		}
		meta["files_changed"] = out.FilesChanged
		meta["additions"] = out.Additions
		meta["removals"] = out.Removals
		meta["changed"] = out.Changed
		meta["git_diff_stored"] = out.GitDiffStored
		meta["full_diff_stored"] = out.FullDiffStored
		meta["config_diff_stored"] = out.ConfigDiffStored
	}

	_ = statusactivities.AwaitPkgStatusUpdateFlowStepStatus(ctx, statusactivities.UpdateStatusRequest{
		ID: s.StepID,
		Status: app.CompositeStatus{
			Status:                 app.StatusSuccess,
			StatusHumanDescription: desc,
			Metadata:               meta,
		},
	})

	logger.Info("run differences step complete",
		"run_id", s.RunID,
		"app_branch_id", s.AppBranchID,
		"skipped", out != nil && out.Skipped)

	return nil
}
