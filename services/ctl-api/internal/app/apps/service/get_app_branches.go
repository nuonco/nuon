package service

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/db/scopes"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

// @ID						GetAppBranches
// @Summary				get app branches
// @Description.markdown	get_app_branches.md
// @Param					app_id						path	string	true	"app ID"
// @Param					offset						query	int		false	"offset of branches to return"	Default(0)
// @Param					limit						query	int		false	"limit of branches to return"	Default(10)
// @Param					page						query	int		false	"page number of results to return"	Default(0)
// @Tags					apps
// @Accept					json
// @Produce				json
// @Security				APIKey && OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	[]app.AppBranch
// @Router					/v1/apps/{app_id}/branches [get]
func (s *service) GetAppBranches(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	enabled, err := s.featuresClient.FeatureEnabled(ctx, app.OrgFeatureAppBranches)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check feature: %w", err))
		return
	}
	if !enabled {
		ctx.Error(features.ErrFeatureNotEnabled(app.OrgFeatureAppBranches))
		return
	}

	appID := ctx.Param("app_id")
	cfgs, err := s.getAppBranches(ctx, org.ID, appID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.JSON(http.StatusOK, cfgs)
}

func (s *service) getAppBranches(ctx *gin.Context, orgID, appID string) ([]app.AppBranch, error) {
	branches := make([]app.AppBranch, 0)

	res := s.db.WithContext(ctx).
		Model(&app.AppBranch{}).
		Select(fmt.Sprintf("app_branches.*, "+
			"(SELECT COUNT(*) FROM %s w "+
			"WHERE w.owner_type = 'app_branches' AND w.owner_id = app_branches.id AND w.deleted_at = 0) AS workflow_count",
			(&app.Workflow{}).TableName())).
		Scopes(scopes.WithOffsetPagination).
		Where(app.AppBranch{
			OrgID: orgID,
			AppID: appID,
		}).
		Order("created_at desc").
		Find(&branches)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get app branches: %w", res.Error)
	}

	branches, err := db.HandlePaginatedResponse(ctx, branches)
	if err != nil {
		return nil, fmt.Errorf("unable to get app branches: %w", err)
	}

	if err := s.attachLatestBranchConfigs(ctx, branches); err != nil {
		return nil, fmt.Errorf("unable to get latest branch configs: %w", err)
	}

	if err := s.attachLatestBranchRuns(ctx, branches); err != nil {
		return nil, fmt.Errorf("unable to get latest branch runs: %w", err)
	}

	return branches, nil
}

func (s *service) attachLatestBranchConfigs(ctx *gin.Context, branches []app.AppBranch) error {
	if len(branches) == 0 {
		return nil
	}

	branchIDs := make([]string, 0, len(branches))
	for _, branch := range branches {
		branchIDs = append(branchIDs, branch.ID)
	}

	latestConfigIDs := s.db.WithContext(ctx).
		Model(&app.AppBranchConfig{}).
		Select("DISTINCT ON (app_branch_id) id").
		Where("app_branch_id IN ?", branchIDs).
		Order("app_branch_id, created_at DESC")

	configs := make([]app.AppBranchConfig, 0)
	res := s.db.WithContext(ctx).
		Preload("ConnectedGithubVCSConfig").
		Preload("PublicGitVCSConfig").
		Preload("InstallGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("\"order\" ASC")
		}).
		Where("id IN (?)", latestConfigIDs).
		Find(&configs)
	if res.Error != nil {
		return res.Error
	}

	configsByBranchID := make(map[string][]app.AppBranchConfig, len(configs))
	for _, config := range configs {
		configsByBranchID[config.AppBranchID] = []app.AppBranchConfig{config}
	}

	for i := range branches {
		branches[i].Configs = configsByBranchID[branches[i].ID]
	}

	return nil
}

func (s *service) attachLatestBranchRuns(ctx *gin.Context, branches []app.AppBranch) error {
	if len(branches) == 0 {
		return nil
	}

	branchIDs := make([]string, 0, len(branches))
	for _, branch := range branches {
		branchIDs = append(branchIDs, branch.ID)
	}

	latestRunIDs := s.db.WithContext(ctx).
		Model(&app.AppBranchRun{}).
		Select("DISTINCT ON (app_branch_id) id").
		Where("app_branch_id IN ?", branchIDs).
		Order("app_branch_id, created_at DESC")

	runs := make([]app.AppBranchRun, 0)
	res := s.db.WithContext(ctx).
		Preload("VCSConnectionCommit").
		Where("id IN (?)", latestRunIDs).
		Find(&runs)
	if res.Error != nil {
		return res.Error
	}

	if err := s.markRunsAwaitingApproval(ctx, runs); err != nil {
		return err
	}

	runsByBranchID := make(map[string]app.AppBranchRun, len(runs))
	for _, run := range runs {
		runsByBranchID[run.AppBranchID] = run
	}

	for i := range branches {
		if run, ok := runsByBranchID[branches[i].ID]; ok {
			branches[i].LatestRun = &run
		}
	}

	return nil
}

func (s *service) markRunsAwaitingApproval(ctx *gin.Context, runs []app.AppBranchRun) error {
	workflowIDs := make([]string, 0, len(runs))
	for _, run := range runs {
		if run.WorkflowID != nil {
			workflowIDs = append(workflowIDs, *run.WorkflowID)
		}
	}
	if len(workflowIDs) == 0 {
		return nil
	}

	awaitingWorkflowIDs := make([]string, 0)
	res := s.db.WithContext(ctx).
		Model(&app.WorkflowStep{}).
		Distinct().
		Joins("JOIN install_workflow_step_approvals approvals "+
			"ON approvals.install_workflow_step_id = install_workflow_steps.id AND approvals.deleted_at = 0").
		Joins("LEFT JOIN install_workflow_step_approval_responses responses "+
			"ON responses.install_workflow_step_approval_id = approvals.id AND responses.deleted_at = 0").
		Where("install_workflow_steps.install_workflow_id IN ?", workflowIDs).
		Where("install_workflow_steps.execution_type = ?", app.WorkflowStepExecutionTypeApproval).
		Where("install_workflow_steps.status->>'status' = ?", string(app.AwaitingApproval)).
		Where("responses.id IS NULL").
		Pluck("install_workflow_steps.install_workflow_id", &awaitingWorkflowIDs)
	if res.Error != nil {
		return res.Error
	}

	awaiting := make(map[string]bool, len(awaitingWorkflowIDs))
	for _, workflowID := range awaitingWorkflowIDs {
		awaiting[workflowID] = true
	}

	for i := range runs {
		if runs[i].WorkflowID != nil && awaiting[*runs[i].WorkflowID] {
			runs[i].AwaitingApproval = true
		}
	}

	return nil
}
