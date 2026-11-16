package alignment

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *Helpers) loadOrigins(ctx context.Context, alignment *InstallAlignment) error {
	workflowIDs := map[string]struct{}{}
	for _, res := range alignment.Resources {
		if res.appliedWorkflowID != "" {
			workflowIDs[res.appliedWorkflowID] = struct{}{}
		}
	}
	workflows := map[string]*app.Workflow{}
	if len(workflowIDs) > 0 {
		var list []app.Workflow
		if err := s.db.WithContext(ctx).Select("id", "type", "metadata").
			Where("id IN ?", generics.MapToKeys(workflowIDs)).Find(&list).Error; err != nil {
			return fmt.Errorf("unable to get alignment origin workflows: %w", err)
		}
		for i := range list {
			workflows[list[i].ID] = &list[i]
		}
	}

	runIDOf := func(workflowID string) string {
		if wf := workflows[workflowID]; wf != nil {
			return generics.FromPtrStr(wf.Metadata["app_branch_run_id"])
		}
		return ""
	}
	runIDs := map[string]struct{}{}
	for id := range workflows {
		if runID := runIDOf(id); runID != "" {
			runIDs[runID] = struct{}{}
		}
	}
	runs := map[string]*InstallOverviewCommit{}
	if len(runIDs) > 0 {
		var list []app.AppBranchRun
		if err := s.db.WithContext(ctx).Preload("VCSConnectionCommit").Where("id IN ?", generics.MapToKeys(runIDs)).Find(&list).Error; err != nil {
			return fmt.Errorf("unable to get alignment branch runs: %w", err)
		}
		for i := range list {
			runs[list[i].ID] = CommitFromRun(&list[i])
		}
	}

	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		if wf := workflows[res.appliedWorkflowID]; wf != nil {
			res.AppliedBy = &InstallAlignmentOrigin{WorkflowID: wf.ID, WorkflowType: wf.Type, Run: runs[runIDOf(wf.ID)]}
		}
	}
	return nil
}
