package helpers

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// FindBaseAppBranchRun returns the most recent deploy run on the same app branch
// with labels.builds_completed=true. Preview runs (git-preview-run, plan-only)
// are excluded from the candidate pool.
func (h *Helpers) FindBaseAppBranchRun(ctx context.Context, appBranchID string) (*app.AppBranchRun, error) {
	var baseRun app.AppBranchRun
	err := h.db.WithContext(ctx).
		Where(app.AppBranchRun{
			AppBranchID: appBranchID,
			Status:      "success",
		}).
		Where(buildsCompletedLabelClause(h.db), "true").
		Where("run_type IN ?", []app.AppBranchRunType{
			app.AppBranchRunTypeGit,
			app.AppBranchRunTypeManual,
		}).
		Where("plan_only = ?", false).
		Order("created_at DESC").
		First(&baseRun).Error
	if err != nil {
		return nil, err
	}
	return &baseRun, nil
}

// buildsCompletedLabelClause returns a dialect-aware SQL fragment for filtering
// AppBranchRun.labels.builds_completed. Postgres uses ->>; sqlite uses json_extract
// so unit tests can exercise the same helper without Postgres.
func buildsCompletedLabelClause(db *gorm.DB) string {
	if db.Dialector.Name() == "sqlite" {
		return "json_extract(labels, '$.builds_completed') = ?"
	}
	return "labels->>'builds_completed' = ?"
}

// shouldCreateComparison reports whether a run of this type gets an AppBranchRunComparison row.
func shouldCreateComparison(runType app.AppBranchRunType, planOnly bool) bool {
	switch runType {
	case app.AppBranchRunTypeGit:
		return true
	case app.AppBranchRunTypeGitPreview:
		return false
	case app.AppBranchRunTypeManual:
		return !planOnly
	default:
		return false
	}
}

// createAppBranchRunComparison creates a comparison row for headRun.
// BaseRunID is set when a prior deploy with builds_completed=true exists; otherwise nil.
// Ownership is HeadRunID on the comparison (has-one from the run); no column on AppBranchRun.
func (h *Helpers) createAppBranchRunComparison(ctx context.Context, headRun *app.AppBranchRun) error {
	var baseRunID *string
	baseRun, err := h.FindBaseAppBranchRun(ctx, headRun.AppBranchID)
	if err == nil {
		baseRunID = &baseRun.ID
	} else if err != gorm.ErrRecordNotFound {
		return err
	}

	comparison := &app.AppBranchRunComparison{
		HeadRunID: headRun.ID,
		BaseRunID: baseRunID,
		OrgID:     headRun.OrgID,
	}
	if err := h.db.WithContext(ctx).Create(comparison).Error; err != nil {
		return err
	}

	headRun.Comparison = comparison
	return nil
}
