package service

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	runsignal "github.com/nuonco/nuon/services/ctl-api/internal/app/apps/signals/branches/run"
	queueclient "github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/client"
)

type AdminTriggerAppBranchRunRequest struct {
	Force bool `json:"force"`
}

// @ID						AdminTriggerAppBranchRun
// @Summary				trigger an app branch run (admin)
// @Description			Admin endpoint to trigger a workflow run for an app branch. Uses the latest config.
// @Tags					apps/admin
// @Accept					json
// @Param					app_branch_id	path	string								true	"app branch ID"
// @Param					req				body	AdminTriggerAppBranchRunRequest	true	"Input"
// @Produce				json
// @Security				AdminEmail
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.AppBranchRun
// @Router					/v1/app-branches/{app_branch_id}/admin-trigger-run [post]
func (s *service) AdminTriggerAppBranchRun(ctx *gin.Context) {
	appBranchID := ctx.Param("app_branch_id")

	var req AdminTriggerAppBranchRunRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		// Allow empty body (default force=false)
		req = AdminTriggerAppBranchRunRequest{}
	}

	// Load branch with queue
	var branch app.AppBranch
	res := s.db.WithContext(ctx).
		Preload("Queue").
		First(&branch, "id = ?", appBranchID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find app branch: %w", res.Error))
		return
	}

	// Get latest config
	var config app.AppBranchConfig
	res = s.db.WithContext(ctx).
		Where("app_branch_id = ?", appBranchID).
		Order("config_number DESC").
		First(&config)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to find latest config for branch: %w", res.Error))
		return
	}

	// Create app branch run
	run, err := s.helpers.CreateAppBranchRun(ctx, &helpers.CreateAppBranchRunRequest{
		AppBranchID:       appBranchID,
		AppBranchConfigID: config.ID,
		Force:             req.Force,
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create app branch run: %w", err))
		return
	}

	// Create workflow
	wf, err := s.helpers.CreateWorkflow(
		ctx,
		appBranchID,
		app.WorkflowTypeAppBranchesRun,
		map[string]string{
			"run_id":        run.ID,
			"config_id":     config.ID,
			"config_number": strconv.Itoa(config.ConfigNumber),
			"force":         strconv.FormatBool(req.Force),
			"event_type":    "manual",
			"commit_sha":    run.CommitSHA,
		},
		false,
	)
	if err != nil {
		run.Status = "failed"
		run.ErrorMessage = fmt.Sprintf("workflow creation failed: %v", err)
		s.db.WithContext(ctx).Save(run)
		ctx.Error(fmt.Errorf("unable to create workflow: %w", err))
		return
	}

	// Update run with workflow ID
	run.WorkflowID = &wf.ID
	res = s.db.WithContext(ctx).Save(run)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to update run with workflow id: %w", res.Error))
		return
	}

	// Enqueue signal
	_, err = s.queueClient.EnqueueSignal(ctx, &queueclient.EnqueueSignalRequest{
		QueueID: branch.Queue.ID,
		Signal: &runsignal.Signal{
			RunID: run.ID,
		},
	})
	if err != nil {
		run.Status = "failed"
		run.ErrorMessage = fmt.Sprintf("signal enqueue failed: %v", err)
		s.db.WithContext(ctx).Save(run)
		ctx.Error(fmt.Errorf("unable to enqueue run signal: %w", err))
		return
	}

	// Reload with relationships
	res = s.db.WithContext(ctx).
		Preload("Workflow").
		Preload("Workflow.Steps").
		Preload("AppBranch").
		Preload("AppBranchConfig").
		First(&run, "id = ?", run.ID)
	if res.Error != nil {
		ctx.Error(fmt.Errorf("unable to reload run: %w", res.Error))
		return
	}

	ctx.JSON(http.StatusCreated, run)
}
