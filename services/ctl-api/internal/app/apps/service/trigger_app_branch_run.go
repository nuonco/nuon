package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	runsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/run"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/features"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type TriggerAppBranchRunRequest struct {
	ConfigID    string `json:"config_id"`     // optional - use latest if not provided
	Force       bool   `json:"force"`         // force run even if no changes detected
	PlanOnly    bool   `json:"plan_only"`     // plan-only preview mode (no apply)
	AppConfigID string `json:"app_config_id"` // optional - use pre-existing app config (skips VCS fetch + config parse)
	SkipBuilds  bool   `json:"skip_builds"`   // skip builds step (e.g. rollback to existing config with existing builds)
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

	// 1. CREATE APP BRANCH RUN FIRST (status=pending, workflow_id=nil)
	run, err := s.helpers.CreateAppBranchRun(ctx, &helpers.CreateAppBranchRunRequest{
		AppBranchID:       appBranchID,
		AppBranchConfigID: config.ID,
		AppConfigID:       req.AppConfigID,
		Force:             req.Force,
		PlanOnly:          req.PlanOnly,
		EventType:         "manual",
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app branch run: %w", err))
		return
	}

	// 2. CREATE WORKFLOW (with run_id in metadata)
	workflowMeta := map[string]string{
		"run_id":        run.ID,
		"app_id":        appID,
		"config_id":     config.ID,
		"config_number": strconv.Itoa(config.ConfigNumber),
		"force":         strconv.FormatBool(req.Force),
		"event_type":    "manual",
		"commit_sha":    run.CommitSHA,
	}
	if req.AppConfigID != "" {
		workflowMeta["app_config_id"] = req.AppConfigID
	}
	if req.SkipBuilds {
		workflowMeta["skip_builds"] = "true"
	}

	workflow, err := s.helpers.CreateWorkflow(
		ctx,
		appBranchID,
		app.WorkflowTypeAppBranchesRun,
		workflowMeta,
		req.PlanOnly,
	)
	if err != nil {
		// Mark run as failed before returning
		run.Status = "failed"
		run.ErrorMessage = fmt.Sprintf("workflow creation failed: %v", err)
		s.db.WithContext(ctx).Save(run)

		ctx.Error(fmt.Errorf("unable to create workflow: %w", err))
		return
	}

	// 3. UPDATE RUN WITH WORKFLOW ID
	run.WorkflowID = &workflow.ID
	res = s.db.WithContext(ctx).Save(run)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to update run with workflow id: %w", res.Error))
		return
	}

	// 4. ENQUEUE SIGNAL and associate the QueueSignal with this run via polymorphic owner
	enqResp, err := s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: branch.Queue.ID,
		Signal: &runsignal.Signal{
			RunID: run.ID,
		},
	})
	if err != nil {
		// Mark run as failed before returning
		run.Status = "failed"
		run.ErrorMessage = fmt.Sprintf("signal enqueue failed: %v", err)
		s.db.WithContext(ctx).Save(run)

		ctx.Error(fmt.Errorf("unable to enqueue run signal: %w", err))
		return
	}
	if res = s.db.WithContext(ctx).Model(&app.QueueSignal{}).
		Where("id = ?", enqResp.ID).
		Updates(map[string]any{
			"owner_id":   run.ID,
			"owner_type": s.db.NamingStrategy.TableName("AppBranchRun"),
		}); res.Error != nil {
		ctx.Error(fmt.Errorf("unable to associate queue signal with run: %w", res.Error))
		return
	}

	// 5. RELOAD RUN WITH RELATIONSHIPS
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

	// 6. RETURN APP BRANCH RUN (not Workflow)
	ctx.JSON(http.StatusCreated, run) // Changed from 'workflow' to 'run'
}
