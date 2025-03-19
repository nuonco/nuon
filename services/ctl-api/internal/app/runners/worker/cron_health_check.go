package worker

import (
	"fmt"
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"
)

const (
	healthCheckWorkflowCronTab string        = "* * * * *"
	heartBeatTimeout           time.Duration = time.Second * 15
	runnerSideCheckInterval    time.Duration = time.Minute * 5
)

// NOTE(fd/sdboyer): we don't want to actively shutdown all runners until this logic is validated
var testOrgIDs = []string{
	"organf4k63tmhyqqgypvukcyty", // prod: Nuon Test V2
	"orgtvkz1podyp9lmenx7o64usx", // prod: Nuon Test
	"orgcz7wndes27hrzcfg5etzohr", // stage
}

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
	// optional: the runner-side healthcheck can be a noop
	Noop bool `json:"noop"`
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

	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to get workflow logger")
	}

	currentStatus, err := activities.AwaitGetRunnerStatusByID(ctx, req.RunnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner status")
	}

	// if we're in a noop status, create a healthcheck and exit
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
	if heartbeat.ID == "" {
		status = "error_fetching_heart_beats"
		return errors.Wrap(err, "unable to get anay heart beats")
	}

	// Ensure the status is created correctly. This will also translate any error that might have
	// occurred while fetching the most recent heartbeat into an appropriate status, and therefore
	// should be done before any other checks.
	newStatus := w.determineStatusFromHeartBeat(ctx, heartbeat)
	status = string(newStatus)

	if newStatus != currentStatus {
		w.mw.Incr(ctx, "runner.health_check_change", metrics.ToTags(tags)...)
		if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          req.RunnerID,
			Status:            newStatus,
			StatusDescription: fmt.Sprintf("status change %s -> %s in health check", currentStatus, newStatus),
		}); err != nil {
			status = "error"
			return errors.Wrap(err, "unable to update runner status")
		}
	}

	healthcheck, err := activities.AwaitCreateHealthCheck(ctx, activities.CreateHealthCheckRequest{
		RunnerID: req.RunnerID,
		Status:   newStatus,
	})
	if err != nil {
		status = "error"
		return errors.Wrap(err, "unable to create runner health check")
	}

	runner, err := activities.AwaitGetByRunnerID(ctx, req.RunnerID)
	if err != nil {
		// This should be unreachable, given that we already retrieved status
		return errors.Wrap(err, "unable to get runner")
	}

	// At this point, we can start executing the individual parts of the Healthcheck as child workflows.
	// These are all child workflows. they return a response w/ a boolean ShouldRestart
	// 1. HealthCheckCheckRestart
	// 2. HealthCheckUpdateNeeded
	// 3. HealthcheckJob: runner-side healthcheck

	// 1. HealthCheckCheckRestart
	hcrreq := &HealthcheckCheckRestartRequest{HeartbeatID: heartbeat.ID, RunnerID: runner.ID}
	hcrres := &HealthcheckCheckRestartResponse{ShouldRestart: false} // default value
	w.execHealthcheckChildWorkflow(ctx, req.RunnerID, "HealthcheckCheckRestart", hcrreq, hcrres)

	// 2. HealthCheckUpdateNeeded
	// If we've got a healthy status and a restart is not already planned, then
	// check to see if the version needs updating.  Only check if healthy, to
	// avoid exacerbating issues the runner may be having. This should also
	// prevent subsequent runs of this workflow from re-attempting the same
	// version check, potentially creating a race condition.

	hcures := &HealthcheckUpdateNeededResponse{} // default value
	if !hcrres.ShouldRestart && newStatus == app.RunnerStatusActive {
		hcureq := &HealthcheckUpdateNeededRequest{HeartbeatID: heartbeat.ID, RunnerID: runner.ID}
		w.execHealthcheckChildWorkflow(ctx, req.RunnerID, "HealthcheckUpdateNeeded", hcureq, hcures)
	}
	// the runner statuse has changed: emit a metric
	if newStatus != currentStatus {
		w.mw.Incr(ctx, "runner.health_check_change", metrics.ToTags(tags)...)
		if err := activities.AwaitUpdateStatus(ctx, activities.UpdateStatusRequest{
			RunnerID:          req.RunnerID,
			Status:            newStatus,
			StatusDescription: fmt.Sprintf("status change %s -> %s in health check", currentStatus, newStatus),
		}); err != nil {
			status = "error"
			return errors.Wrap(err, "unable to update runner status")
		}
	}

	// 3. HealthcheckJob: runner-side healthcheck
	// TODO(fd): determine a cadence - we probably don't want to run this every single time we run the healthcheck
	hcjres := &HealthcheckJobRunnerResponse{ShouldRestart: false} // default value
	// TODO(fd): disabled for now until we validate everything works
	// if startTS.Minute()%5 == 0 {                                  // NOTE(fd): make the right side of the % a configurable var or at least a constant
	// 	// As a child workflow
	// 	hcjreq := HealthcheckJobRunnerRequest{HealthCheckID: healthcheck.ID, RunnerID: runner.ID}
	// 	w.execHealthcheckChildWorkflow(ctx, req.RunnerID, "HealthcheckJobRunner", hcjreq, hcjres)
	// }

	// use the responses to determine if we need to restart the runner
	if hcures.ShouldUpdate {
		// send update job
		l.Info("runner should be updated",
			zap.Bool("HealthcheckCheckRestart.ShouldRestart", hcrres.ShouldRestart),
			zap.Bool("HealthcheckUpdateNeeded.ShouldUpdate", hcures.ShouldUpdate),
			zap.Bool("HealthcheckJob.ShouldRestart", hcjres.ShouldRestart),
		)
		// NOTE(fd/sdboyer): not enabled at this time
		// w.evClient.Send(ctx, runner.ID, &signals.Signal{
		// 	Type:          signals.OperationUpdateVersion,
		// 	HealthCheckID: healthcheck.ID,
		// })
	} else if hcrres.ShouldRestart {
		// graceful shutdown
		l.Info("runner should be restarted",
			zap.String("strategy", "graceful"),
			zap.Bool("HealthcheckCheckRestart.ShouldRestart", hcrres.ShouldRestart),
			zap.Bool("HealthcheckUpdateNeeded.ShouldUpdate", hcures.ShouldUpdate),
			zap.Bool("HealthcheckJob.ShouldRestart", hcjres.ShouldRestart),
		)
		if generics.SliceContains(runner.OrgID, testOrgIDs) {
			// only shutdown specific subset of runners
			w.gracefulShutdown(ctx, startTS, l, runner, healthcheck)
		}
	} else if hcjres.ShouldRestart {
		// forceful shutdown
		l.Info("runner should be restarted",
			zap.String("strategy", "forceful"),
			zap.Bool("HealthcheckCheckRestart.ShouldRestart", hcrres.ShouldRestart),
			zap.Bool("HealthcheckUpdateNeeded.ShouldUpdate", hcures.ShouldUpdate),
			zap.Bool("HealthcheckJob.ShouldRestart", hcjres.ShouldRestart),
		)
		// TODO(fd): implement forceful shutdown w/ AwaitExecuteJob
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

func (w *Workflows) gracefulShutdown(ctx workflow.Context, startTS time.Time, l *zap.Logger, runner *app.Runner, healthcheck *app.RunnerHealthCheck) error {
	// NOTE(fd): this method could be inlined - it's separated for clarity, not re-use
	// 1. only send out every 5th minute. this is due to the fact it would otherwise be possible to send out shutdown jobs before
	// the runner got a chance to get up and running trapping a runner in a shutdown loop.
	// 2. only send out a signal if there isn't currently another shutdown job in the queue. otherwise we'll allow these to pile up
	// which would lead to a situation where a runner starts up only to process a queue of shutdown jobs preventing it from ever
	// becoming healthy.

	shutdownJobs, err := activities.AwaitGetRunnerShutdownJobQueueByRunnerID(ctx, runner.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner shutdown job queue")
	}
	shutdownJobIDs := []string{}
	for _, sj := range shutdownJobs {
		shutdownJobIDs = append(shutdownJobIDs, sj.ID)
	}

	if startTS.Minute()%5 != 0 { // only send every 3rd minute
		l.Debug(
			"refusing to send shutdown signal - time is not right",
			zap.Any("shutdown_job_ids", shutdownJobIDs),
		)
		return nil
	}
	if len(shutdownJobs) > 0 { // do not send if there are other/existing shut down jobs
		l.Warn(
			"refusing to send shutdown signal - shutdown jobs exist in queue",
			zap.Any("shutdown_job_ids", shutdownJobIDs),
		)
		return nil
	}

	l.Debug("sending shutdown signal")
	w.evClient.Send(ctx, runner.ID, &signals.Signal{
		Type:          signals.OperationShutdown,
		HealthCheckID: healthcheck.ID,
	})
	return nil
}
