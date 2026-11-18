package syncstatus

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

func (s *Helpers) loadOrigins(ctx context.Context, orgID string, syncStatus *InstallResourceSyncStatus) error {
	runIDs := map[string]struct{}{}
	for _, res := range syncStatus.Resources {
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
			return fmt.Errorf("unable to get sync status branch runs: %w", err)
		}
		for i := range list {
			runs[list[i].ID] = CommitFromRun(&list[i])
		}
	}

	for i := range syncStatus.Resources {
		res := &syncStatus.Resources[i]
		if res.appliedWorkflowID != "" && res.appliedWorkflow != "" {
			res.AppliedBy = &InstallResourceSyncStatusOrigin{WorkflowID: res.appliedWorkflowID, WorkflowType: res.appliedWorkflow, Run: runs[res.appliedRunID]}
		}
	}
	return nil
}
