package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
)

type UpdateRunnerProcessRequest struct {
	Status            app.RunnerProcessStatus `json:"status" validate:"required" swaggertype:"string"`
	StatusDescription string                  `json:"status_description"`
}

// @ID						UpdateRunnerProcess
// @Summary				update a runner process
// @Description.markdown	update_runner_process.md
// @Param					req			body	UpdateRunnerProcessRequest	true	"Input"
// @Param					runner_id	path	string						true	"runner ID"
// @Param					process_id	path	string						true	"process ID"
// @Tags					runners/runner
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.RunnerProcess
// @Router					/v1/runners/{runner_id}/processes/{process_id} [PATCH]
func (s *service) UpdateRunnerProcess(ctx *gin.Context) {
	processID := ctx.Param("process_id")

	var req UpdateRunnerProcessRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.Error(stderr.NewInvalidRequest(err))
		return
	}

	process, err := s.updateRunnerProcess(ctx, processID, req)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to update runner process: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, process)
}

func (s *service) updateRunnerProcess(ctx context.Context, processID string, req UpdateRunnerProcessRequest) (*app.RunnerProcess, error) {
	// get current process for composite status history
	current, err := s.getRunnerProcess(ctx, processID)
	if err != nil {
		return nil, err
	}

	// build new composite status with history
	newComposite := app.NewCompositeStatus(ctx, app.Status(req.Status))
	newComposite.StatusHumanDescription = req.StatusDescription
	newComposite.History = append([]app.CompositeStatus{current.CompositeStatus}, current.CompositeStatus.History...)
	// flatten: keep only top-level history
	newComposite.History[0].History = nil

	updates := app.RunnerProcess{
		CompositeStatus: newComposite,
	}

	res := s.db.WithContext(ctx).
		Model(&app.RunnerProcess{ID: processID}).
		Updates(updates)
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update runner process: %w", res.Error)
	}

	return s.getRunnerProcess(ctx, processID)
}
