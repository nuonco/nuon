package service

import (
	"context"
	"fmt"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/alignment"
)

type InstallAlignment = alignment.InstallAlignment
type InstallAlignmentResource = alignment.InstallAlignmentResource
type InstallAlignmentResourceKind = alignment.InstallAlignmentResourceKind
type InstallAlignmentCategory = alignment.InstallAlignmentCategory
type InstallAlignmentAttempt = alignment.InstallAlignmentAttempt
type InstallAlignmentOpenDeployment = alignment.InstallAlignmentOpenDeployment
type InstallAlignmentExplanation = alignment.InstallAlignmentExplanation
type InstallAlignmentPlan = alignment.InstallAlignmentPlan

const (
	InstallAlignmentResourceKindComponent = alignment.InstallAlignmentResourceKindComponent
	InstallAlignmentResourceKindImage     = alignment.InstallAlignmentResourceKindImage
	InstallAlignmentResourceKindSandbox   = alignment.InstallAlignmentResourceKindSandbox
	InstallAlignmentResourceKindStack     = alignment.InstallAlignmentResourceKindStack

	InstallAlignmentCategoryChangesNotApplied  = alignment.InstallAlignmentCategoryChangesNotApplied
	InstallAlignmentCategoryStateDriftDetected = alignment.InstallAlignmentCategoryStateDriftDetected
	InstallAlignmentCategoryInSync             = alignment.InstallAlignmentCategoryInSync
	InstallAlignmentCategoryUnknown            = alignment.InstallAlignmentCategoryUnknown
	InstallAlignmentCategoryDisabled           = alignment.InstallAlignmentCategoryDisabled
	InstallAlignmentCategoryTornDown           = alignment.InstallAlignmentCategoryTornDown
)

func (s *service) alignmentBranchRunCounts(ctx context.Context, install *app.Install, branchID string, alignment *InstallAlignment) error {
	if branchID == "" {
		return nil
	}
	configIDs := []string{}
	for _, res := range alignment.Resources {
		if res.LastAttempt != nil && res.LastAttempt.BranchRun != nil {
			configIDs = append(configIDs, res.LastAttempt.BranchRun.AppConfigID)
		}
	}
	if len(configIDs) == 0 {
		return nil
	}
	var counts []struct {
		ID    string
		Count int
	}
	if err := s.db.WithContext(ctx).Model(&app.AppConfig{}).
		Select("app_configs.id, COUNT(DISTINCT newer.id) AS count").
		Joins(`LEFT JOIN app_configs newer ON newer.org_id = app_configs.org_id
			AND newer.app_id = app_configs.app_id AND newer.deleted_at = 0
			AND (newer.created_at > app_configs.created_at OR (newer.created_at = app_configs.created_at AND newer.id > app_configs.id))
			AND newer.id IN (?)`, installTrackingRuns(s.db.WithContext(ctx), install.OrgID, branchID).Select("app_config_id")).
		Where(app.AppConfig{OrgID: install.OrgID, AppID: install.AppID}).
		Where("app_configs.id IN ?", configIDs).
		Where("app_configs.id IN (?)", installTrackingRuns(s.db.WithContext(ctx), install.OrgID, branchID).Select("app_config_id")).
		Group("app_configs.id").
		Scan(&counts).Error; err != nil {
		return fmt.Errorf("unable to count newer app branch runs: %w", err)
	}
	byConfig := map[string]*int{}
	for _, count := range counts {
		byConfig[count.ID] = &count.Count
	}
	for i := range alignment.Resources {
		res := &alignment.Resources[i]
		if res.LastAttempt != nil && res.LastAttempt.BranchRun != nil {
			res.NewerBranchRuns = byConfig[res.LastAttempt.BranchRun.AppConfigID]
		}
	}
	return nil
}
