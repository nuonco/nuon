package worker

import (
	"fmt"
	"strconv"
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

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
	status := "ok"
	tags := map[string]string{}
	defer func() {
		tags["status"] = status
		e2eLatency := workflow.Now(ctx).Sub(startTS)
		w.mw.Incr(ctx, "runner.health_check", metrics.ToTags(tags)...)
		w.mw.Timing(ctx, "runner.health_check_timing", e2eLatency, metrics.ToTags(tags)...)
	}()

	currentStatus, err := activities.AwaitGetRunnerStatusByID(ctx, req.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner status")
	}

	noopStatus := generics.SliceContains(currentStatus, []app.RunnerStatus{
		app.RunnerStatusProvisioning,
		app.RunnerStatusDeprovisioning,
		app.RunnerStatusReprovisioning,
		app.RunnerStatusDeprovisioned,
	})
	if noopStatus {
		_, err := activities.AwaitCreateHealthCheck(ctx, activities.CreateHealthCheckRequest{
			RunnerID: req.RunnerID,
			Status:   currentStatus,
		})
		if err != nil {
			status = "error"
			return errors.Wrap(err, "unable to create runner health check")
		}
		return nil
	}

	// The next few checks require the most recent heartbeat. Fetch it, then pass it down to them
	heartbeat, err := activities.AwaitGetMostRecentHeartBeatRequestByRunnerID(ctx, req.RunnerID)
	if err != nil {
		status = "error_fetching_heart_beats"
		return errors.Wrap(err, "unable to get status from heart beats")
	}

	// Ensure the status is created correctly. This will also translate any error that might have
	// occurred while fetching the most recent heartbeat into an appropriate status, and therefore
	// should be done before any other checks.
	newStatus := w.determineStatusFromHeartBeat(ctx, heartbeat)
	status = string(newStatus)

	_, err = activities.AwaitCreateHealthCheck(ctx, activities.CreateHealthCheckRequest{
		RunnerID: req.RunnerID,
		Status:   newStatus,
	})
	if err != nil {
		status = "error"
		return errors.Wrap(err, "unable to create runner health check")
	}

	if err := w.checkRecentRestart(ctx, heartbeat, req.RunnerID); err != nil {
		return errors.Wrap(err, "unable to check recent restart")
	}

	// If we've got a healthy status, then check to see if the version needs
	// updating.  Only check if healthy, to avoid exacerbating issues the runner
	// may be having. This should also prevent subsequent runs of this workflow
	// from re-attempting the same version check, potentially creating a race
	// condition.
	if newStatus == app.RunnerStatusActive {
		// err = w.checkUpdateNeeded(ctx, heartbeat, healthcheck, req.RunnerID)
		if err != nil {
			status = "error_failed_update_check"
			return errors.Wrap(err, "failed to check for needed update")
		}
	}

	// no change as the old status is the same as the new status
	if newStatus == currentStatus {
		return nil
	}

	// the runner changed statuses
	w.mw.Incr(ctx, "runner.health_check_change", metrics.ToTags(tags)...)
	if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
		RunnerID:          req.RunnerID,
		Status:            newStatus,
		StatusDescription: fmt.Sprintf("status change %s -> %s in health check", currentStatus, newStatus),
	}); err != nil {
		status = "error"
		return errors.Wrap(err, "unable to update runner status")
	}

	return nil
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

func (w *Workflows) checkRecentRestart(
	ctx workflow.Context,
	heartbeat *app.RunnerHeartBeat,
	runnerID string,
) error {
	if heartbeat.AliveTime > time.Second*5 {
		return nil
	}

	runner, err := activities.AwaitGetByRunnerID(ctx, runnerID)
	if err != nil {
		// This should be unreachable, given that we already retrieved status
		return errors.Wrap(err, "unable to get runner")
	}
	w.mw.Incr(ctx, "runner.restart", metrics.ToTags(map[string]string{
		"runner_type": string(runner.RunnerGroup.Type),
	})...)

	return nil
}

func (w *Workflows) checkUpdateNeeded(
	ctx workflow.Context,
	heartbeat *app.RunnerHeartBeat,
	healthcheck *app.RunnerHealthCheck,
	runnerID string,
) error {
	runner, err := activities.AwaitGetByRunnerID(ctx, runnerID)
	if err != nil {
		// This should be unreachable, given that we already retrieved status
		return errors.Wrap(err, "unable to get runner")
	}

	var needsUpdate bool
	if runner.RunnerGroup.Settings.ExpectedVersion == "latest" {
		needsUpdate = heartbeat.Version != w.cfg.Version
	} else if heartbeat.Version != runner.RunnerGroup.Settings.ExpectedVersion {
		// NOTE(sdboyer) this branch is unreachable until we have a versioning
		// strategy other than latest.
		needsUpdate = true
	}

	// NOTE(jm): if we need an update, we just write a metric
	w.mw.Incr(ctx, "runner.version_update", metrics.ToTags(map[string]string{
		"runner_type":          string(runner.RunnerGroup.Type),
		"needs_version_update": strconv.FormatBool(needsUpdate),
		"expected_latest":      strconv.FormatBool(runner.RunnerGroup.Settings.ExpectedVersion == "latest"),
	})...)

	if needsUpdate {
		l, err := log.WorkflowLogger(ctx)
		if err != nil {
			return nil
		}
		l.Info("sending signal to update out-of-date runner",
			zap.String("runner_id", runnerID),
			zap.String("runner_type", string(runner.RunnerGroup.Type)),
			zap.String("expected_version", runner.RunnerGroup.Settings.ExpectedVersion),
			zap.String("reported_version", heartbeat.Version),
			zap.String("api_version", w.cfg.Version),
		)

		//w.evClient.Send(ctx, runnerID, &signals.RequestSignal{
		//Signal: &signals.Signal{
		//Type:          signals.OperationUpdateVersion,
		//HealthCheckID: healthcheck.ID,
		//},
		//EventLoopRequest: eventloop.EventLoopRequest{
		//ID: runnerID,
		//},
		//})
	}

	return nil
}
