package alignment

import (
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *Helpers) loadOrigins(ctx context.Context, st *installAlignmentState, alignment *InstallAlignment) error {
	var version app.InstallAppConfigVersion
	if st.install.AppConfigID != "" {
		err := s.db.WithContext(ctx).
			Select("id", "app_branch_run_id", "workflow_id").
			Where(app.InstallAppConfigVersion{InstallID: st.install.ID, NewAppConfigID: st.install.AppConfigID}).
			Order("app_branch_run_id IS NULL, workflow_id IS NULL, created_at DESC").
			First(&version).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("unable to get selected app config version: %w", err)
		}
	}
	expectedWorkflowID := generics.FromPtrStr(version.WorkflowID)

	workflowIDs := map[string]struct{}{}
	if expectedWorkflowID != "" {
		workflowIDs[expectedWorkflowID] = struct{}{}
	}
	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		if res.appliedWorkflowID != "" {
			workflowIDs[res.appliedWorkflowID] = struct{}{}
		}
		if res.BlockingAttempt != nil && res.BlockingAttempt.WorkflowID != "" {
			workflowIDs[res.BlockingAttempt.WorkflowID] = struct{}{}
		}
	}
	workflows := map[string]*app.Workflow{}
	if len(workflowIDs) > 0 {
		ids := make([]string, 0, len(workflowIDs))
		for id := range workflowIDs {
			ids = append(ids, id)
		}
		var list []app.Workflow
		if err := s.db.WithContext(ctx).Select("id", "type", "metadata").Where("id IN ?", ids).Find(&list).Error; err != nil {
			return fmt.Errorf("unable to get alignment workflows: %w", err)
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
	expectedRunID := generics.FromPtrStr(version.AppBranchRunID)
	if expectedRunID == "" {
		expectedRunID = runIDOf(expectedWorkflowID)
	}
	runIDs := map[string]struct{}{}
	if expectedRunID != "" {
		runIDs[expectedRunID] = struct{}{}
	}
	for id := range workflows {
		if runID := runIDOf(id); runID != "" {
			runIDs[runID] = struct{}{}
		}
	}
	runs := map[string]*InstallOverviewCommit{}
	if len(runIDs) > 0 {
		ids := make([]string, 0, len(runIDs))
		for id := range runIDs {
			ids = append(ids, id)
		}
		var list []app.AppBranchRun
		if err := s.db.WithContext(ctx).Preload("VCSConnectionCommit").Where("id IN ?", ids).Find(&list).Error; err != nil {
			return fmt.Errorf("unable to get alignment branch runs: %w", err)
		}
		for i := range list {
			runs[list[i].ID] = CommitFromRun(&list[i])
		}
	}

	origin := func(workflowID, runID string) *InstallAlignmentOrigin {
		if workflowID == "" && runID == "" {
			return nil
		}
		o := &InstallAlignmentOrigin{WorkflowID: workflowID, Run: runs[runID]}
		if wf := workflows[workflowID]; wf != nil {
			o.WorkflowType = wf.Type
		}
		return o
	}
	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		res.AppliedBy = origin(res.appliedWorkflowID, runIDOf(res.appliedWorkflowID))
		if res.ExpectedConfigID != "" && res.Category != InstallAlignmentCategoryInSync {
			res.ExpectedBy = origin(expectedWorkflowID, expectedRunID)
		}
		if res.BlockingAttempt != nil {
			res.BlockingAttempt.Run = runs[runIDOf(res.BlockingAttempt.WorkflowID)]
		}
	}
	return nil
}
