package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals/processjob"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type AdminCreateInstallRunnerShutDownJobRequest struct{}

// @ID						AdminCreateInstallRunnerShutDownJob
// @Summary				shut down a runners by install ID
// @Description.markdown	shut_down_runner_by_install_id.md
// @Param					install_id	path	string										true	"install ID"
// @Param					req			body	AdminCreateInstallRunnerShutDownJobRequest	true	"Input"
// @Tags					runners/admin
// @Security				AdminEmail
// @Accept					json
// @Produce				json
// @Success				201	{boolean}	true
// @Router					/v1/installs/{install_id}/runners/shutdown-job [POST]
func (s *service) AdminCreateInstallRunnerqShutDownJob(ctx *gin.Context) {
	installID := ctx.Param("install_id")

	var req AdminCreateInstallRunnerShutDownJobRequest
	if err := ctx.ShouldBindJSON(&req); err != nil && !errors.Is(err, io.EOF) {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	install, err := s.adminGetInstall(ctx, installID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get install %s: %w", installID, err))
		return
	}

	for _, runner := range install.RunnerGroup.Runners {
		job, err := s.adminCreateJob(ctx, runner.ID, app.RunnerJobTypeShutDown)
		if err != nil {
			ctx.Error(fmt.Errorf("unable to create shutdown job: %w", err))
			return
		}

		if err := s.helpers.EnqueueRunnerSignal(ctx, runner.ID, &processjob.Signal{RunnerID: runner.ID, JobID: job.ID}); err != nil {
			ctx.Error(fmt.Errorf("unable to enqueue process-job signal: %w", err))
			return
		}
	}

	ctx.JSON(http.StatusCreated, true)
}

func (s *service) adminGetInstall(ctx context.Context, installID string) (*app.Install, error) {
	install := app.Install{}
	res := s.db.WithContext(ctx).
		Preload("AWSAccount").
		Preload("AzureAccount").
		Preload("GCPAccount").
		Preload("App").
		Preload("App.Org").
		Preload("CreatedBy").
		Preload("InstallInputs").
		Preload("RunnerGroup").
		Preload("RunnerGroup.Settings").
		Preload("RunnerGroup.Runners").
		Preload("InstallSandboxRuns", func(db *gorm.DB) *gorm.DB {
			return db.Order("install_sandbox_runs.created_at DESC")
		}).
		Preload("InstallSandboxRuns.AppSandboxConfig").
		Or("id = ?", installID).
		First(&install)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to get install: %w", res.Error)
	}

	return &install, nil
}
