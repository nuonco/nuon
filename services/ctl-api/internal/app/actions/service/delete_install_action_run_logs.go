package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"slices"
	"time"

	chgo "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/nuonco/nuon/pkg/generics"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/middlewares/stderr"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/audit"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
)

const deleteActionRunLogsTimeout = time.Minute

var deletableActionRunStatuses = []app.InstallActionWorkflowRunStatus{
	app.InstallActionRunStatusFinished,
	app.InstallActionRunStatusError,
	app.InstallActionRunStatusTimedOut,
	app.InstallActionRunStatusCancelled,
	app.InstallActionRunStatusRetried,
}

// @ID						DeleteInstallActionRunLogs
// @Summary				delete the logs and trace spans of an action run
// @Description.markdown	delete_install_action_run_logs.md
// @Param					install_id	path	string	true	"install ID"
// @Param					run_id		path	string	true	"run ID"
// @Tags					actions
// @Accept					json
// @Produce				json
// @Security				APIKey
// @Security				OrgID
// @Failure				400	{object}	stderr.ErrResponse
// @Failure				401	{object}	stderr.ErrResponse
// @Failure				403	{object}	stderr.ErrResponse
// @Failure				404	{object}	stderr.ErrResponse
// @Failure				500	{object}	stderr.ErrResponse
// @Success				200	{object}	app.LogStream
// @Router					/v1/installs/{install_id}/actions/runs/{run_id}/logs [DELETE]
func (s *service) DeleteInstallActionRunLogs(ctx *gin.Context) {
	org, err := cctx.OrgFromContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	acct, err := cctx.AccountFromGinContext(ctx)
	if err != nil {
		ctx.Error(err)
		return
	}

	if !s.accountsHelpers.IsOrgAdmin(acct, org.ID) {
		ctx.Error(stderr.ErrAuthorization{
			Err:         fmt.Errorf("only org admins can delete action run logs"),
			Description: "only org admins can delete action run logs",
		})
		return
	}

	installID := ctx.Param("install_id")
	runID := ctx.Param("run_id")

	run := app.InstallActionWorkflowRun{}
	res := s.db.WithContext(ctx).
		Preload("LogStream").
		Where(app.InstallActionWorkflowRun{ID: runID, OrgID: org.ID, InstallID: installID}).
		First(&run)
	if res.Error != nil {
		if errors.Is(res.Error, gorm.ErrRecordNotFound) {
			ctx.Error(stderr.ErrNotFound{Err: res.Error, Description: "action run not found"})
			return
		}
		ctx.Error(fmt.Errorf("unable to get action run: %w", res.Error))
		return
	}

	logStream := run.LogStream
	if logStream.ID == "" {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("action run %s has no log stream", runID),
			Description: "this action run has no logs",
		})
		return
	}

	if !slices.Contains(deletableActionRunStatuses, run.Status) {
		ctx.Error(stderr.ErrUser{
			Err:         fmt.Errorf("action run %s is still running", runID),
			Description: "logs can only be deleted after the action run finishes",
		})
		return
	}

	if err := s.deleteActionRunLogs(ctx, org.ID, &logStream); err != nil {
		ctx.Error(err)
		return
	}

	if logStream.LogsDeletedAt == nil {
		now := time.Now()
		res = s.db.WithContext(ctx).
			Model(&logStream).
			Where("logs_deleted_at IS NULL").
			Updates(app.LogStream{LogsDeletedAt: &now, LogsDeletedByID: acct.ID})
		if res.Error != nil {
			ctx.Error(fmt.Errorf("unable to mark log stream as deleted: %w", res.Error))
			return
		}

		res = s.db.WithContext(ctx).
			Where(app.LogStream{ID: logStream.ID}).
			First(&logStream)
		if res.Error != nil {
			ctx.Error(fmt.Errorf("unable to reload log stream: %w", res.Error))
			return
		}
	}

	s.audit.Emit(ctx, audit.Event{
		Type:        audit.EventInstallActionWorkflowRun,
		Message:     "action run logs deleted",
		Outcome:     audit.OutcomeSucceeded,
		InstallID:   installID,
		SubjectID:   run.ID,
		SubjectType: "install_action_workflow_runs",
		Attrs: map[string]string{
			"action_workflow_run.id": run.ID,
			"log_stream.id":          logStream.ID,
			"deleted_by.id":          acct.ID,
		},
	})

	ctx.JSON(http.StatusOK, logStream)
}

func (s *service) deleteActionRunLogs(ctx context.Context, orgID string, logStream *app.LogStream) error {
	var jobIDs []string
	res := s.db.WithContext(ctx).
		Model(&app.RunnerJob{}).
		Where(app.RunnerJob{LogStreamID: generics.ToPtr(logStream.ID)}).
		Pluck("id", &jobIDs)
	if res.Error != nil {
		return fmt.Errorf("unable to get runner jobs for log stream: %w", res.Error)
	}

	ctx, cancel := context.WithTimeout(ctx, deleteActionRunLogsTimeout)
	defer cancel()

	chCtx := chgo.Context(ctx, chgo.WithSettings(chgo.Settings{
		"mutations_sync":     2,
		"max_execution_time": int(deleteActionRunLogsTimeout.Seconds()),
	}))

	res = s.chDB.WithContext(chCtx).
		Unscoped().
		Where(app.OtelLogRecord{OrgID: orgID, LogStreamID: logStream.ID}).
		Delete(&app.OtelLogRecord{})
	if res.Error != nil {
		return fmt.Errorf("unable to delete log stream logs: %w", res.Error)
	}

	if len(jobIDs) == 0 {
		return nil
	}

	if !logStream.ParentLogStreamID.Empty() {
		res = s.chDB.WithContext(chCtx).
			Unscoped().
			Where(app.OtelLogRecord{OrgID: orgID, LogStreamID: logStream.ParentLogStreamID.String}).
			Where("runner_job_id IN ?", jobIDs).
			Delete(&app.OtelLogRecord{})
		if res.Error != nil {
			return fmt.Errorf("unable to delete parent log stream logs: %w", res.Error)
		}
	}

	res = s.chDB.WithContext(chCtx).
		Unscoped().
		Where("runner_job_id IN ?", jobIDs).
		Delete(&app.OtelTraceIngestion{})
	if res.Error != nil {
		return fmt.Errorf("unable to delete trace spans: %w", res.Error)
	}

	return nil
}
