package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	runsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/v2/branches/run"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type TriggerAppBranchRunRequest struct {
	ConfigID string `json:"config_id"` // optional - use latest if not provided
	Force    bool   `json:"force"`     // force run even if no changes detected
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

	// Feature flag checks
	if !org.Features[string(app.OrgFeatureAppBranches)] {
		ctx.Error(fmt.Errorf("app branches feature not enabled for this organization"))
		return
	}

	if !org.Features[string(app.OrgFeatureQueues)] {
		ctx.Error(fmt.Errorf("queues feature not enabled for this organization"))
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
		Force:             req.Force,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app branch run: %w", err))
		return
	}

	// 2. CREATE WORKFLOW (with run_id in metadata)
	workflow, err := s.helpers.CreateWorkflow(
		ctx,
		appBranchID,
		app.WorkflowTypeAppBranchesRun,
		map[string]string{
			"run_id":        run.ID, // NEW: Include run ID
			"config_id":     config.ID,
			"config_number": strconv.Itoa(config.ConfigNumber),
			"force":         strconv.FormatBool(req.Force),
		},
		false, // not plan only
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

	// 4. ENQUEUE SIMPLIFIED SIGNAL (just run_id)
	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: branch.Queue.ID,
		Signal: &runsignal.Signal{
			RunID: run.ID, // SIMPLIFIED: only pass run_id
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
	ctx.Header(app.HeaderInstallWorkflowID, workflow.ID)
	ctx.JSON(http.StatusCreated, run) // Changed from 'workflow' to 'run'
}
