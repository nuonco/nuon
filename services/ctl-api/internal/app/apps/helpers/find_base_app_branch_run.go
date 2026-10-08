package helpers

import (
	"context"

	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/plugins"
)

// FindBaseAppBranchRun returns the most recent non-preview run on the app
// branch that parsed an app config and finished component builds. The run's
// overall status is not required to be success: a later deploy failure still
// leaves a usable baseline.
func (h *Helpers) FindBaseAppBranchRun(ctx context.Context, appBranchID string) (*app.AppBranchRun, error) {
	return h.findLatestBaseRun(ctx, appBranchID, "")
}

// FindBaseAppBranchRunForHead resolves preview baselines from the git branch a
// pull request targets. Regular branch runs compare with the previous parsed
// build on the same app branch.
func (h *Helpers) FindBaseAppBranchRunForHead(ctx context.Context, headRun *app.AppBranchRun) (*app.AppBranchRun, error) {
	if headRun == nil {
		return nil, gorm.ErrRecordNotFound
	}
	if headRun.BaseBranch == "" || (headRun.RunType != app.AppBranchRunTypeGitPreview && headRun.Preview == nil) {
		return h.FindBaseAppBranchRun(ctx, headRun.AppBranchID)
	}

	targetBranchID, err := h.findTargetAppBranchID(ctx, headRun)
	if err != nil {
		return nil, err
	}

	return h.findLatestBaseRun(ctx, targetBranchID, headRun.ID)
}

func (h *Helpers) findLatestBaseRun(ctx context.Context, appBranchID, excludedRunID string) (*app.AppBranchRun, error) {
	query := h.db.WithContext(ctx).
		Where(app.AppBranchRun{AppBranchID: appBranchID}).
		Where("app_config_id <> ''").
		Where("labels->>'builds_completed' = ?", "true").
		Where("run_type IN ?", []app.AppBranchRunType{
			app.AppBranchRunTypeGit,
			app.AppBranchRunTypeManual,
		}).
		Where("plan_only = ?", false)
	if excludedRunID != "" {
		query = query.Not(app.AppBranchRun{ID: excludedRunID})
	}

	var baseRun app.AppBranchRun
	err := query.
		Order("created_at DESC").
		First(&baseRun).Error
	if err != nil {
		return nil, err
	}
	return &baseRun, nil
}

func (h *Helpers) findTargetAppBranchID(ctx context.Context, headRun *app.AppBranchRun) (string, error) {
	var headBranch app.AppBranch
	if err := h.db.WithContext(ctx).First(&headBranch, "id = ?", headRun.AppBranchID).Error; err != nil {
		return "", err
	}

	appBranchConfigTable := plugins.TableName(h.db, app.AppBranchConfig{})
	var target struct {
		AppBranchID string
	}
	err := h.db.WithContext(ctx).
		Table("connected_github_vcs_configs").
		Select("app_branch_configs.app_branch_id").
		Joins("JOIN app_branch_configs ON app_branch_configs.id = connected_github_vcs_configs.component_config_id AND connected_github_vcs_configs.component_config_type = ?", appBranchConfigTable).
		Joins("JOIN app_branches ON app_branches.id = app_branch_configs.app_branch_id AND app_branches.deleted_at = 0").
		Where("app_branches.app_id = ?", headBranch.AppID).
		Where("connected_github_vcs_configs.branch = ?", headRun.BaseBranch).
		Where("connected_github_vcs_configs.deleted_at = 0").
		Where("app_branch_configs.deleted_at = 0").
		Order("app_branch_configs.created_at DESC").
		Limit(1).
		Scan(&target).Error
	if err != nil {
		return "", err
	}
	if target.AppBranchID != "" {
		return target.AppBranchID, nil
	}

	err = h.db.WithContext(ctx).
		Table("public_git_vcs_configs").
		Select("app_branch_configs.app_branch_id").
		Joins("JOIN app_branch_configs ON app_branch_configs.id = public_git_vcs_configs.component_config_id AND public_git_vcs_configs.component_config_type = ?", appBranchConfigTable).
		Joins("JOIN app_branches ON app_branches.id = app_branch_configs.app_branch_id AND app_branches.deleted_at = 0").
		Where("app_branches.app_id = ?", headBranch.AppID).
		Where("public_git_vcs_configs.branch = ?", headRun.BaseBranch).
		Where("public_git_vcs_configs.deleted_at = 0").
		Where("app_branch_configs.deleted_at = 0").
		Order("app_branch_configs.created_at DESC").
		Limit(1).
		Scan(&target).Error
	if err != nil {
		return "", err
	}
	if target.AppBranchID != "" {
		return target.AppBranchID, nil
	}

	var branch app.AppBranch
	if err := h.db.WithContext(ctx).
		Where(app.AppBranch{AppID: headBranch.AppID, Name: headRun.BaseBranch}).
		First(&branch).Error; err != nil {
		return "", err
	}
	return branch.ID, nil
}

// shouldCreateComparison reports whether a run of this type gets an AppBranchRunComparison row.
func shouldCreateComparison(runType app.AppBranchRunType, planOnly bool) bool {
	switch runType {
	case app.AppBranchRunTypeGit:
		return true
	case app.AppBranchRunTypeGitPreview:
		return true
	case app.AppBranchRunTypeManual:
		return !planOnly
	default:
		return false
	}
}

// createAppBranchRunComparison creates a comparison row for headRun.
// BaseRunID is set when a prior run parsed an app config and finished builds; otherwise nil.
// Ownership is HeadRunID on the comparison (has-one from the run); no column on AppBranchRun.
func (h *Helpers) createAppBranchRunComparison(ctx context.Context, headRun *app.AppBranchRun) error {
	var baseRunID *string
	baseRun, err := h.FindBaseAppBranchRunForHead(ctx, headRun)
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
