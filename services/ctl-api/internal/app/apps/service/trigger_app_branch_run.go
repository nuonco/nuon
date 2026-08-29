package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
)

type PreviewRunRequest struct {
	Source   app.AppBranchRunPreviewSource `json:"source"`
	PRNumber *int                          `json:"pr_number,omitempty"`
	GitRef   string                        `json:"git_ref,omitempty"`
	HeadSHA  string                        `json:"head_sha,omitempty"`

	Mode      *app.AppBranchRunPreviewMode `json:"mode,omitempty"`
	InstallID *string                      `json:"install_id,omitempty"`
}

type TriggerAppBranchRunRequest struct {
	ConfigID    string `json:"config_id"`
	Force       bool   `json:"force"`
	PlanOnly    bool   `json:"plan_only"`
	AppConfigID string `json:"app_config_id"`
	SkipBuilds  bool   `json:"skip_builds"`

	SyncAppConfig bool `json:"sync_app_config"`

	AutoApprove bool `json:"auto_approve"`

	PRNumber   *int   `json:"pr_number,omitempty"`
	HeadSHA    string `json:"head_sha,omitempty"`
	BaseBranch string `json:"base_branch,omitempty"`

	PreviewRun *PreviewRunRequest `json:"preview_run,omitempty"`
}

func (c *TriggerAppBranchRunRequest) Validate(v *validator.Validate) error {
	if err := v.Struct(c); err != nil {
		return err
	}
	if c.SyncAppConfig && c.AppConfigID == "" {
		return fmt.Errorf("sync_app_config requires app_config_id")
	}
	if c.SyncAppConfig && c.SkipBuilds {
		return fmt.Errorf("sync_app_config cannot be combined with skip_builds: a freshly synced config has no builds yet")
	}
	if c.PreviewRun != nil {
		if !c.PreviewRun.Source.Valid() || c.PreviewRun.Source == "" {
			return fmt.Errorf("preview_run.source is required")
		}
		if c.PlanOnly {
			return fmt.Errorf("plan_only cannot be combined with preview_run")
		}
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
		ctx.Error(stderr.NewInvalidRequest(err))
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
	if req.SyncAppConfig {
		workflowMeta["sync_app_config"] = "true"
	}
	// fetchcommit only honours HeadSHA on a git-preview run.
	runType := app.AppBranchRunTypeManual
	planOnly := req.PlanOnly
	prNumber := req.PRNumber
	headSHA := req.HeadSHA
	baseBranch := req.BaseBranch

	var previewInput *helpers.PreviewRunInput
	if req.PreviewRun != nil {
		previewInput = &helpers.PreviewRunInput{
			Source:           req.PreviewRun.Source,
			PRNumber:         req.PreviewRun.PRNumber,
			GitRef:           req.PreviewRun.GitRef,
			HeadSHA:          req.PreviewRun.HeadSHA,
			InputAppConfigID: req.AppConfigID,
		}
		if req.PreviewRun.Mode != nil || req.PreviewRun.InstallID != nil {
			previewInput.Override = &app.AppBranchPreviewOverride{
				Mode:      req.PreviewRun.Mode,
				InstallID: req.PreviewRun.InstallID,
			}
		}
		if previewInput.PRNumber != nil {
			prNumber = previewInput.PRNumber
		}
		if previewInput.HeadSHA != "" {
			headSHA = previewInput.HeadSHA
		}
		switch previewInput.Source {
		case app.AppBranchRunPreviewSourcePR, app.AppBranchRunPreviewSourceCommit, app.AppBranchRunPreviewSourceBranch:
			runType = app.AppBranchRunTypeGitPreview
		case app.AppBranchRunPreviewSourceLocal:
			if req.AppConfigID == "" {
				ctx.Error(stderr.NewInvalidRequest(fmt.Errorf("preview_run with source local requires app_config_id")))
				return
			}
		}
	} else if req.PlanOnly && (req.PRNumber != nil || req.HeadSHA != "") {
		runType = app.AppBranchRunTypeGitPreview
	}

	if prNumber != nil {
		workflowMeta["pr_number"] = strconv.Itoa(*prNumber)
	}
	if headSHA != "" {
		workflowMeta["head_sha"] = headSHA
	}
	if baseBranch != "" {
		workflowMeta["base_branch"] = baseBranch
	}

	approvalOption := app.InstallApprovalOptionApproveAll
	if !req.AutoApprove {
		approvalOption, err = s.helpers.ResolveAppBranchApprovalOption(ctx, appID, appBranchID, config.ID)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to resolve approval option: %w", err))
			return
		}
	}

	triggerResp, err := s.helpers.TriggerAppBranchRun(ctx, &helpers.TriggerAppBranchRunRequest{
		Run: helpers.CreateAppBranchRunRequest{
			AppBranchID:       appBranchID,
			AppBranchConfigID: config.ID,
			AppConfigID:       req.AppConfigID,
			Force:             req.Force,
			PlanOnly:          planOnly,
			RunType:           runType,
			EventType:         "manual",
			PRNumber:          prNumber,
			HeadSHA:           headSHA,
			BaseBranch:        baseBranch,
			Preview:           previewInput,
		},
		QueueID:        branch.Queue.ID,
		Metadata:       workflowMeta,
		ApprovalOption: approvalOption,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to trigger app branch run: %w", err))
		return
	}
	run := triggerResp.Run

	// The standalone sync path advanced this in appconfigsync's finalize step,
	// which a branch-run sync never reaches.
	if req.SyncAppConfig {
		if acct, acctErr := cctx.AccountFromContext(ctx); acctErr == nil {
			if journeyErr := s.accountsHelpers.UpdateUserJourneyStepForFirstAppSync(ctx, acct.ID, appID); journeyErr != nil {
				s.l.Warn("unable to update app_synced journey step", zap.Error(journeyErr))
			}
		}
	}

	res = s.db.WithContext(ctx).
		Preload("Workflow").
		Preload("Workflow.Steps").
		Preload("Workflow.CreatedBy").
		Preload("AppBranch").
		Preload("AppBranchConfig").
		Preload("CreatedBy").
		Preload("Preview").
		First(&run, "id = ?", run.ID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to reload run: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusCreated, run)
}
