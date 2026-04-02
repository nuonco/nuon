package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/runners/signals"
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

	s.evClient.Send(ctx, runnerID, &signals.Signal{
		Type:         signals.OperationProcessShutdown,
		ProcessID:    processID,
		ShutdownType: string(req.ShutdownType),
	})

	ctx.JSON(http.StatusCreated, shutdown)
}

func (s *service) createRunnerProcessShutdown(ctx context.Context, processID string, req ShutdownRunnerProcessRequest) (*app.RunnerProcessShutdown, error) {
	shutdown := app.RunnerProcessShutdown{
		RunnerProcessID: processID,
		Type:            req.ShutdownType,
		Status:          app.RunnerProcessShutdownStatusRequested,
		CompositeStatus: app.NewCompositeStatus(ctx, app.Status(app.RunnerProcessShutdownStatusRequested)),
	}

	if res := s.db.WithContext(ctx).Create(&shutdown); res.Error != nil {
		return nil, fmt.Errorf("unable to create runner process shutdown: %w", res.Error)
	}

	return &shutdown, nil
}
