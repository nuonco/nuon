package service

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const appBranchWorkflowOwnerType = "app_branches"

// @ID						GetAppBranchRun
// @Summary					get an app branch workflow run
// @Description			Returns a branch workflow by either app branch run ID or workflow ID.
// @Tags					apps
// @Param					app_id			path	string	true	"app ID"
// @Param					app_branch_id	path	string	true	"app branch ID"
// @Param					run_id			path	string	true	"app branch run ID or workflow ID"
// @Accept					json
// @Produce					json
// @Security				APIKey
// @Security				OrgID
// @Failure					400	{object}	stderr.ErrResponse
// @Failure					401	{object}	stderr.ErrResponse
// @Failure					403	{object}	stderr.ErrResponse
// @Failure					404	{object}	stderr.ErrResponse
// @Failure					500	{object}	stderr.ErrResponse
// @Success					200	{object}	app.Workflow
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/runs/{run_id} [get]
func (s *service) GetAppBranchRun(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get org from context: %w", err))
		return
	}

	appID := ctx.Param("app_id")
	appBranchID := ctx.Param("app_branch_id")
	runID := ctx.Param("run_id")

	var branch app.AppBranch
	if err := s.db.WithContext(ctx).
		Where(app.AppBranch{
			ID:    appBranchID,
			AppID: appID,
			OrgID: org.ID,
		}).
		First(&branch).Error; err != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", err))
		return
	}

	workflowID, err := s.resolveAppBranchWorkflowID(ctx, branch.ID, runID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to find app branch run: %w", err))
		return
	}

	workflow, err := s.getAppBranchWorkflow(ctx, org.ID, branch.ID, workflowID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get app branch workflow: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, workflow)
}

func (s *service) resolveAppBranchWorkflowID(ctx *gin.Context, appBranchID, runID string) (string, error) {
	var run app.AppBranchRun
	err := s.db.WithContext(ctx).
		Select("workflow_id").
		Where(app.AppBranchRun{
			ID:          runID,
			AppBranchID: appBranchID,
		}).
		First(&run).Error
	if err == nil {
		if run.WorkflowID == nil || *run.WorkflowID == "" {
			return "", gorm.ErrRecordNotFound
		}
		return *run.WorkflowID, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", err
	}

	var workflow app.Workflow
	err = s.db.WithContext(ctx).
		Select("id").
		Where(app.Workflow{
			ID:        runID,
			OwnerID:   appBranchID,
			OwnerType: appBranchWorkflowOwnerType,
		}).
		First(&workflow).Error
	if err != nil {
		return "", err
	}
	return workflow.ID, nil
}

func (s *service) getAppBranchWorkflow(ctx *gin.Context, orgID, appBranchID, workflowID string) (*app.Workflow, error) {
	var workflow app.Workflow
	err := s.db.WithContext(ctx).
		Preload("CreatedBy").
		Preload("Steps", func(db *gorm.DB) *gorm.DB {
			return db.Order("group_idx, group_retry_idx, idx, created_at asc")
		}).
		Preload("Steps.CreatedBy").
		Preload("Steps.Approval", func(db *gorm.DB) *gorm.DB {
			return db.Omit("contents")
		}).
		Preload("Steps.Approval.Response").
		Preload("AppBranchRuns").
		Preload("AppBranchRuns.VCSConnectionCommit").
		Preload("AppBranchRuns.AppBranchConfig").
		Preload("AppBranchRuns.AppBranchConfig.ConnectedGithubVCSConfig").
		Preload("AppBranchRuns.AppBranchConfig.PublicGitVCSConfig").
		Preload("AppBranchRuns.Preview").
		Preload("StepGroups", func(db *gorm.DB) *gorm.DB {
			return db.Order("group_idx asc")
		}).
		Where(app.Workflow{
			ID:        workflowID,
			OrgID:     orgID,
			OwnerID:   appBranchID,
			OwnerType: appBranchWorkflowOwnerType,
		}).
		First(&workflow).Error
	if err != nil {
		return nil, err
	}
	return &workflow, nil
}
