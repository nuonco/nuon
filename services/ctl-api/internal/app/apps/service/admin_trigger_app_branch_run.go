package service

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/apps/helpers"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
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
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
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

	triggerResp, err := s.helpers.TriggerAppBranchRun(ctx, &helpers.TriggerAppBranchRunRequest{
		Run: helpers.CreateAppBranchRunRequest{
			AppBranchID:       appBranchID,
			AppBranchConfigID: config.ID,
			Force:             req.Force,
		},
		QueueID: branch.Queue.ID,
		Metadata: map[string]string{
			"config_id":     config.ID,
			"config_number": strconv.Itoa(config.ConfigNumber),
			"force":         strconv.FormatBool(req.Force),
			"event_type":    "manual",
		},
	})
	if err != nil {
		ctx.Error(fmt.Errorf("unable to trigger app branch run: %w", err))
		return
	}
	run := triggerResp.Run

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
