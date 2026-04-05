package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @ID						CompleteRunnerProcessShutdown
// @Summary				mark a runner process shutdown as completed
// @Description.markdown	complete_runner_process_shutdown.md
// @Param					runner_id	path	string	true	"runner ID"
// @Param					process_id	path	string	true	"process ID"
// @Param					shutdown_id	path	string	true	"shutdown ID"
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
// @Success				200	{object}	app.RunnerProcessShutdown
// @Router					/v1/runners/{runner_id}/processes/{process_id}/shutdowns/{shutdown_id}/complete [POST]
func (s *service) CompleteRunnerProcessShutdown(ctx *gin.Context) {
	processID := ctx.Param("process_id")
	shutdownID := ctx.Param("shutdown_id")

	shutdown, err := s.completeRunnerProcessShutdown(ctx, processID, shutdownID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to complete runner process shutdown: %w", err))
		return
	}

	ctx.JSON(http.StatusOK, shutdown)
}

func (s *service) completeRunnerProcessShutdown(ctx context.Context, processID, shutdownID string) (*app.RunnerProcessShutdown, error) {
	var shutdown app.RunnerProcessShutdown
	if res := s.db.WithContext(ctx).First(&shutdown, "id = ? AND runner_process_id = ?", shutdownID, processID); res.Error != nil {
		return nil, fmt.Errorf("unable to find shutdown: %w", res.Error)
	}

	// Mark the shutdown record as completed
	newComposite := app.NewCompositeStatus(ctx, app.Status(app.RunnerProcessShutdownStatusCompleted))
	newComposite.StatusHumanDescription = "shutdown completed by runner"
	newComposite.History = append([]app.CompositeStatus{shutdown.CompositeStatus}, shutdown.CompositeStatus.History...)
	newComposite.History[0].History = nil

	res := s.db.WithContext(ctx).
		Model(&app.RunnerProcessShutdown{ID: shutdownID}).
		Updates(app.RunnerProcessShutdown{
			CompositeStatus: newComposite,
		})
	if res.Error != nil {
		return nil, fmt.Errorf("unable to update shutdown status: %w", res.Error)
	}

	// Transition the process to shut-down so the process_shutdown signal
	// workflow can detect it and complete cleanly.
	s.updateProcessStatusShutDown(ctx, processID)

	var updated app.RunnerProcessShutdown
	if res := s.db.WithContext(ctx).First(&updated, "id = ?", shutdownID); res.Error != nil {
		return nil, fmt.Errorf("unable to get updated shutdown: %w", res.Error)
	}

	return &updated, nil
}

func (s *service) updateProcessStatusShutDown(ctx context.Context, processID string) {
	var process app.RunnerProcess
	if res := s.db.WithContext(ctx).First(&process, "id = ?", processID); res.Error != nil {
		s.l.Warn("unable to get process for shut-down transition", zap.String("process_id", processID), zap.Error(res.Error))
		return
	}

	newComposite := app.NewCompositeStatus(ctx, app.Status(app.RunnerProcessStatusShutDown))
	newComposite.StatusHumanDescription = "shutdown completed"
	newComposite.History = append([]app.CompositeStatus{process.CompositeStatus}, process.CompositeStatus.History...)
	newComposite.History[0].History = nil

	if res := s.db.WithContext(ctx).
		Model(&app.RunnerProcess{ID: processID}).
		Updates(app.RunnerProcess{
			CompositeStatus: newComposite,
		}); res.Error != nil {
		s.l.Warn("unable to update process to shut-down", zap.String("process_id", processID), zap.Error(res.Error))
	}
}
