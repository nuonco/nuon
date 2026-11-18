package alignment

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *Helpers) loadOrigins(ctx context.Context, orgID string, alignment *InstallAlignment) error {
	runIDs := map[string]struct{}{}
	for _, res := range alignment.Resources {
		if res.appliedRunID != "" {
			runIDs[res.appliedRunID] = struct{}{}
		}
	}
	runs := map[string]*InstallOverviewCommit{}
	if len(runIDs) > 0 {
		var list []app.AppBranchRun
		if err := s.db.WithContext(ctx).Preload("VCSConnectionCommit").
			Where(app.AppBranchRun{OrgID: orgID}).
			Where("id IN ?", generics.MapToKeys(runIDs)).
			Find(&list).Error; err != nil {
			return fmt.Errorf("unable to get alignment branch runs: %w", err)
		}
		for i := range list {
			runs[list[i].ID] = CommitFromRun(&list[i])
		}
	}

	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		if res.appliedWorkflowID != "" && res.appliedWorkflow != "" {
			res.AppliedBy = &InstallAlignmentOrigin{WorkflowID: res.appliedWorkflowID, WorkflowType: res.appliedWorkflow, Run: runs[res.appliedRunID]}
		}
	}
	return nil
}
