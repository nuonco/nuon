package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type TriggerAppBranchRunRequest struct {
	ConfigID    string `json:"config_id"`     // optional - use latest if not provided
	Force       bool   `json:"force"`         // force run even if no changes detected
	PlanOnly    bool   `json:"plan_only"`     // plan-only preview mode (no apply)
	AppConfigID string `json:"app_config_id"` // optional - use pre-existing app config (skips VCS fetch + config parse)
	SkipBuilds  bool   `json:"skip_builds"`   // skip builds step (e.g. rollback to existing config with existing builds)

	// PR context, for previews triggered from CI rather than a GitHub webhook.
	// Supplying PRNumber is what lets the run report back onto the pull request.
	PRNumber   *int   `json:"pr_number,omitempty"`
	HeadSHA    string `json:"head_sha,omitempty"`
	BaseBranch string `json:"base_branch,omitempty"`
}

func (c *TriggerAppBranchRunRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return err
	}
	return nil
}

// @ID						TriggerAppBranchRun
// @Summary				trigger app branch workflow run
// @Description			Creates and triggers a workflow run for an app branch. If config_id is not provided, uses the latest config.
// @Tags					apps
// @Accept					json
// @Param					req				body	TriggerAppBranchRunRequest	true	"Input"
// @Param					app_id			path	string						true	"app ID"
// @Param					app_branch_id	path	string						true	"app branch ID"
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppBranchRun
// @Router					/v1/apps/{app_id}/branches/{app_branch_id}/runs [post]
func (s *service) TriggerAppBranchRun(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	enabled, err := s.featuresClient.AllFeaturesEnabled(ctx, app.OrgFeatureAppBranches)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to check features: %w", err))
		return
	}
	if !enabled {
		ctx.Error(features.ErrFeatureNotEnabled(app.OrgFeatureAppBranches))
		return
	}

	appID := ctx.Param("app_id")
	appBranchID := ctx.Param("app_branch_id")

	var req TriggerAppBranchRunRequest
	if err := ctx.BindJSON(&req); err != nil {
		ctx.Error(fmt.Errorf("unable to parse request: %w", err))
		return
	}
	if err := req.Validate(s.v); err != nil {
		ctx.Error(fmt.Errorf("invalid request: %w", err))
		return
	}

	// Verify branch exists and belongs to this org/app
	var branch app.AppBranch
	res := s.db.WithContext(ctx).
		Preload("Queue").
		Where(app.AppBranch{
			OrgID: org.ID,
			AppID: appID,
		}).
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", res.Error))
		return
	}

	// Validate app_config_id if provided
	if req.AppConfigID != "" {
		var appCfg app.AppConfig
		res = s.db.WithContext(ctx).
			Where(app.AppConfig{
				AppID: appID,
				OrgID: org.ID,
			}).
			First(&appCfg, "id = ?", req.AppConfigID)
		if res.Error != nil {
			ctx.Error(fmt.Errorf("unable to find app config: %w", res.Error))
			return
		}
	}

	// Load config (by ID or latest)
	var config app.AppBranchConfig
	if req.ConfigID != "" {
		res = s.db.WithContext(ctx).
			Where("app_branch_id = ?", appBranchID).
			First(&config, "id = ?", req.ConfigID)
		if res.Error != nil {
			ctx.Error(fmt.Errorf("unable to find config: %w", res.Error))
			return
		}
	} else {
		// Get latest config
		res = s.db.WithContext(ctx).
			Where("app_branch_id = ?", appBranchID).
			Order("config_number DESC").
			First(&config)
		if res.Error != nil {
			ctx.Error(fmt.Errorf("unable to find latest config: %w", res.Error))
			return
		}
	}

	workflowMeta := map[string]string{
		"app_id":        appID,
		"config_id":     config.ID,
		"config_number": strconv.Itoa(config.ConfigNumber),
		"force":         strconv.FormatBool(req.Force),
		"event_type":    "manual",
	}
	if req.AppConfigID != "" {
		workflowMeta["app_config_id"] = req.AppConfigID
	}
	if req.SkipBuilds {
		workflowMeta["skip_builds"] = "true"
	}
	if req.PRNumber != nil {
		workflowMeta["pr_number"] = strconv.Itoa(*req.PRNumber)
	}
	if req.HeadSHA != "" {
		workflowMeta["head_sha"] = req.HeadSHA
	}
	if req.BaseBranch != "" {
		workflowMeta["base_branch"] = req.BaseBranch
	}

	triggerResp, err := s.helpers.TriggerAppBranchRun(ctx, &helpers.TriggerAppBranchRunRequest{
		Run: helpers.CreateAppBranchRunRequest{
			AppBranchID:       appBranchID,
			AppBranchConfigID: config.ID,
			AppConfigID:       req.AppConfigID,
			Force:             req.Force,
			PlanOnly:          req.PlanOnly,
			EventType:         "manual",
			PRNumber:          req.PRNumber,
			HeadSHA:           req.HeadSHA,
			BaseBranch:        req.BaseBranch,
		},
		QueueID:  branch.Queue.ID,
		Metadata: workflowMeta,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to trigger app branch run: %w", err))
		return
	}
	run := triggerResp.Run

	res = s.db.WithContext(ctx).
		Preload("Workflow").
		Preload("Workflow.Steps").
		Preload("Workflow.CreatedBy").
		Preload("AppBranch").
		Preload("AppBranchConfig").
		Preload("CreatedBy").
		First(&run, "id = ?", run.ID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to reload run: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusCreated, run)
}
