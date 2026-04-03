package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

type ShutdownRunnerProcessRequest struct {
	ShutdownType app.RunnerProcessShutdownType `json:"shutdown_type" validate:"required" swaggertype:"string"`
}

// @ID						ShutdownRunnerProcess
// @Summary				request shutdown of a runner process
// @Description.markdown	shutdown_runner_process.md
// @Param					req			body	ShutdownRunnerProcessRequest	true	"Input"
// @Param					runner_id	path	string							true	"runner ID"
// @Param					process_id	path	string							true	"process ID"
// @Tags					runners
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				201	{object}	app.RunnerProcessShutdown
// @Router					/v1/runners/{runner_id}/processes/{process_id}/shutdown [POST]
func (s *service) ShutdownRunnerProcess(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	runnerID := ctx.Param("runner_id")
	processID := ctx.Param("process_id")

	var req ShutdownRunnerProcessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	// verify the process belongs to this runner and org
	process, err := s.getRunnerProcess(ctx, processID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to get runner process: %w", err))
		return
	}
	if process.RunnerID != runnerID || process.OrgID != org.ID {
		ctx.Error(fmt.Errorf("runner process not found"))
		return
	}

	shutdown, err := s.createRunnerProcessShutdown(ctx, processID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to create runner process shutdown: %w", err))
		return
	}

	// Immediately mark process as pending-shutdown so health checks noop
	if err := s.updateProcessStatusPendingShutdown(ctx, process); err != nil {
		s.l.Warn("unable to set process pending-shutdown status", zap.Error(err))
	}

	// Write a red health check to ClickHouse so dashboards reflect the shutdown
	s.createShutdownHealthCheck(ctx, process.RunnerID, processID)

	// Enqueue shutdown signal to the v2 process queue and stop health check emitters
	if err := s.helpers.EnqueueProcessShutdown(ctx, runnerID, processID, req.ShutdownType); err != nil {
		s.l.Warn("unable to enqueue process shutdown signal", zap.Error(err))
	}

	ctx.JSON(http.StatusCreated, shutdown)
}

func (s *service) updateProcessStatusPendingShutdown(ctx context.Context, process *app.RunnerProcess) error {
	newComposite := app.NewCompositeStatus(ctx, app.Status(app.RunnerProcessStatusPendingShutdown))
	newComposite.StatusHumanDescription = "Shutdown pending"
	newComposite.History = append([]app.CompositeStatus{process.CompositeStatus}, process.CompositeStatus.History...)
	newComposite.History[0].History = nil

	res := s.db.WithContext(ctx).
		Model(&app.RunnerProcess{ID: process.ID}).
		Updates(app.RunnerProcess{
			CompositeStatus: newComposite,
		})
	if res.Error != nil {
		return fmt.Errorf("unable to update process status: %w", res.Error)
	}

	return nil
}

func (s *service) createRunnerProcessShutdown(ctx context.Context, processID string, req ShutdownRunnerProcessRequest) (*app.RunnerProcessShutdown, error) {
	shutdown := app.RunnerProcessShutdown{
		RunnerProcessID: processID,
		Type:            req.ShutdownType,
		CompositeStatus: app.NewCompositeStatus(ctx, app.Status(app.RunnerProcessShutdownStatusRequested)),
	}

	if res := s.db.WithContext(ctx).Create(&shutdown); res.Error != nil {
		return nil, fmt.Errorf("unable to create runner process shutdown: %w", res.Error)
	}

	return &shutdown, nil
}
