package service

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/metrics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
)

// @ID						ReportRunnerProcessTerminating
// @Summary				report that a runner process is terminating because its host VM is shutting down
// @Description.markdown	report_runner_process_terminating.md
// @Param					runner_id	path	string	true	"runner ID"
// @Param					process_id	path	string	true	"process ID"
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
// @Success				202	{object}	app.RunnerProcess
// @Router					/v1/runners/{runner_id}/processes/{process_id}/terminating [POST]
func (s *service) ReportRunnerProcessTerminating(ctx *gin.Context) {
	runnerID := ctx.Param("runner_id")
	processID := ctx.Param("process_id")

	process, err := s.reportRunnerProcessTerminating(ctx, runnerID, processID)
	if err != nil {
		ctx.Error(fmt.Errorf("unable to report runner process terminating: %w", err))
		return
	}

	ctx.JSON(http.StatusAccepted, process)
}

// reportRunnerProcessTerminating records a best-effort "my host VM is shutting
// down" beacon sent by the runner when it observes an OS shutdown. The runner
// only knows the VM is going away; attribution lives here: if Nuon issued the
// shutdown there is an open RunnerProcessShutdown row, otherwise the
// termination was initiated externally (e.g. the customer stopped/terminated
// the VM from their cloud portal).
func (s *service) reportRunnerProcessTerminating(ctx context.Context, runnerID, processID string) (*app.RunnerProcess, error) {
	var process app.RunnerProcess
	if res := s.db.WithContext(ctx).
		Where(app.RunnerProcess{ID: processID, RunnerID: runnerID}).
		First(&process); res.Error != nil {
		return nil, fmt.Errorf("unable to find runner process: %w", res.Error)
	}

	reason := "external"
	if s.hasOpenProcessShutdown(ctx, processID) {
		reason = "nuon"
	}

	newComposite := app.NewCompositeStatus(ctx, app.Status(app.RunnerProcessStatusShuttingDown))
	newComposite.StatusHumanDescription = fmt.Sprintf("vm terminating (%s)", reason)
	// Carry prior metadata forward and stamp the attribution. Subsequent
	// status transitions (UpdateRunnerProcessStatus) copy Metadata forward, so
	// termination_reason rides onto the live status through the eventual
	// offline/inactive transition — letting RunnerProcess.AfterQuery surface a
	// durable indicator without scanning history.
	for k, v := range process.CompositeStatus.Metadata {
		newComposite.Metadata[k] = v
	}
	newComposite.Metadata["termination_reason"] = reason
	newComposite.History = append([]app.CompositeStatus{process.CompositeStatus}, process.CompositeStatus.History...)
	newComposite.History[0].History = nil

	if res := s.db.WithContext(ctx).
		Model(&app.RunnerProcess{ID: processID}).
		Updates(app.RunnerProcess{CompositeStatus: newComposite}); res.Error != nil {
		return nil, fmt.Errorf("unable to update process status: %w", res.Error)
	}

	s.l.Info("runner process reported terminating",
		zap.String("runner_id", runnerID),
		zap.String("process_id", processID),
		zap.String("org_id", process.OrgID),
		zap.String("reason", reason),
	)
	s.mw.Incr("runner.process.terminating", metrics.ToTags(map[string]string{
		"runner_id": runnerID,
		"org_id":    process.OrgID,
		"reason":    reason,
	}))

	var updated app.RunnerProcess
	if res := s.db.WithContext(ctx).
		Where(app.RunnerProcess{ID: processID}).
		First(&updated); res.Error != nil {
		return nil, fmt.Errorf("unable to get updated process: %w", res.Error)
	}
	return &updated, nil
}

// hasOpenProcessShutdown reports whether Nuon has an in-flight shutdown for the
// process (requested or in-progress). A match means the termination was
// Nuon-initiated rather than external.
func (s *service) hasOpenProcessShutdown(ctx context.Context, processID string) bool {
	var shutdowns []app.RunnerProcessShutdown
	if res := s.db.WithContext(ctx).
		Where(app.RunnerProcessShutdown{RunnerProcessID: processID}).
		Find(&shutdowns); res.Error != nil {
		s.l.Warn("unable to query process shutdowns for terminating attribution",
			zap.String("process_id", processID), zap.Error(res.Error))
		return false
	}

	for _, sd := range shutdowns {
		switch sd.Status {
		case app.RunnerProcessShutdownStatusRequested, app.RunnerProcessShutdownStatusInProgress:
			return true
		}
	}
	return false
}
