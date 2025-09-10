package worker

import (
	"fmt"
	"strconv"
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

const (
	healthCheckWorkflowCronTab string        = "* * * * *"
	heartBeatTimeout           time.Duration = time.Second * 15
	runnerSideCheckInterval    time.Duration = time.Minute * 5
)

func healthCheckWorkflowID(runnerID string) string {
	return fmt.Sprintf("health-check-%s", runnerID)
}

func (w *Workflows) startHealthCheckWorkflow(ctx workflow.Context, req HealthCheckRequest) {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            healthCheckWorkflowID(req.RunnerID),
		CronSchedule:          healthCheckWorkflowCronTab,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, w.HealthCheck, req)
}

// Run a cron to check the health of a runner
type HealthCheckRequest struct {
	RunnerID string `validate:"required" json:"runner_id"`
}

func (w *Workflows) HealthCheck(ctx workflow.Context, req *HealthCheckRequest) error {
	startTS := workflow.Now(ctx)
	runnerStatus := "active"
	healthCheckStatus := "ok"
	changed := false
	runnerType := "unknown"

	defer func() {
		tags := metrics.ToTags(map[string]string{
			"health_check_status":   healthCheckStatus,
			"runner_status":         runnerStatus,
			"runner_status_changed": strconv.FormatBool(changed),
			"runner_type":           runnerType,
		})
		// write metrics now
		w.mw.Incr(ctx, "runner.health_check", tags...)
		w.mw.Timing(ctx, "runner.health_check.latency", time.Now().Sub(startTS), tags...)
	}()

	runner, err := activities.AwaitGetByRunnerID(ctx, req.RunnerID)
	if err != nil {
		healthCheckStatus = "unable_to_get_runner"
		return errors.Wrap(err, "unable to get runner by id")
	}
	runnerType = string(runner.RunnerGroup.Type)

	noopHealthCheck, err := w.isNoopHealthCheck(ctx, req.RunnerID)
	if err != nil {
		healthCheckStatus = "unable_to_check_noop"
		return errors.Wrap(err, "unable to check if a noop health check")
	}
	if noopHealthCheck {
		healthCheckStatus = "noop"
		runnerStatus = "noop"
		return nil
	}

	newStatus, statusChanged, err := w.executeHealthCheck(ctx, req.RunnerID)
	if err != nil {
		healthCheckStatus = "unable_to_execute"
		return errors.Wrap(err, "unable to execute health check")
	}
	runnerStatus = string(newStatus)
	changed = statusChanged

	return nil
}

func (w *Workflows) isNoopHealthCheck(ctx workflow.Context, runnerID string) (bool, error) {
	runner, err := activities.AwaitGetByRunnerID(ctx, runnerID)
	if err != nil {
		return false, errors.Wrap(err, "unable to get runner")
	}

	isNoop := generics.SliceContains(runner.Status, []app.RunnerStatus{
		app.RunnerStatusPending,
		app.RunnerStatusProvisioning,
		app.RunnerStatusDeprovisioning,
		app.RunnerStatusReprovisioning,
		app.RunnerStatusDeprovisioned,
		app.RunnerStatusOffline,
		app.RunnerStatusAwaitingInstallStackRun,
	})
	return isNoop, nil
}

func (w *Workflows) executeHealthCheck(ctx workflow.Context, runnerID string) (app.RunnerStatus, bool, error) {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return app.RunnerStatusUnknown, false, errors.Wrap(err, "unable to get logger")
	}

	runner, err := activities.AwaitGetByRunnerID(ctx, runnerID)
	if err != nil {
		return app.RunnerStatusUnknown, false, errors.Wrap(err, "unable to get runner")
	}

	newStatus := app.RunnerStatusActive
	heartbeat, err := activities.AwaitGetMostRecentHeartBeatRequestByRunnerID(ctx, runnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			newStatus = app.RunnerStatusError
		}

		return app.RunnerStatusUnknown, false, nil
	}
	if heartbeat == nil {
		newStatus = app.RunnerStatusError
	}

	minHeartBeatTS := workflow.Now(ctx).Add(-heartBeatTimeout)
	if heartbeat.CreatedAt.Before(minHeartBeatTS) {
		newStatus = app.RunnerStatusError
	}

	isChanged := runner.Status != newStatus

	_, err = activities.AwaitCreateHealthCheck(ctx, activities.CreateHealthCheckRequest{
		RunnerID: runnerID,
		Status:   newStatus,
	})
	if err != nil {
		return app.RunnerStatusUnknown, false, errors.Wrap(err, "unable to create runner health check")
	}

	if isChanged {
		if newStatus != app.RunnerStatusActive {
			l.Error("runner became unhealthy",
				zap.String("runner_id", runner.ID),
				zap.String("org_id", runner.OrgID),
				zap.String("org_name", runner.Org.Name),
				zap.Int("org_priority", runner.Org.Priority),
			)
		}

		if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          runnerID,
			Status:            newStatus,
			StatusDescription: fmt.Sprintf("status change %s -> %s in health check", runner.Status, newStatus),
		}); err != nil {
			return app.RunnerStatusUnknown, false, errors.Wrap(err, "unable to update runner status")
		}
	}

	return newStatus, isChanged, nil
}

func (w *Workflows) determineStatusFromHeartBeat(ctx workflow.Context, heartbeat *app.RunnerHeartBeat) app.RunnerStatus {
	if heartbeat == nil {
		return app.RunnerStatusError
	}

	minHeartBeatTS := workflow.Now(ctx).Add(-heartBeatTimeout)
	if heartbeat.CreatedAt.Before(minHeartBeatTS) {
		return app.RunnerStatusError
	}

	return app.RunnerStatusActive
}
