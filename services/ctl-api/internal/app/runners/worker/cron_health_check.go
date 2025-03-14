package worker

import (
	"fmt"
	"strconv"
	"time"

	"github.com/DataDog/datadog-go/v5/statsd"
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
)

// NOTE(sdboyer): temporary list of non-customer runners in prod that we're restricting auto-restart logic to as part of testing
var testRunners = []string{
	"runqyk4e77l4qlvh32763u9ovt", // org runner for Nuon Test V2 org
	"run2xty256i7ej2bitai717mxx", // install runner for Telluride Industries app within Nuon Test V2 org
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

	restarting, err := w.checkRestart(ctx, heartbeat, healthcheck, runner)
	if err != nil {
		return errors.Wrap(err, "unable to check recent restart")
	}

	// If we've got a healthy status and a restart is not already planned, then
	// check to see if the version needs updating.  Only check if healthy, to
	// avoid exacerbating issues the runner may be having. This should also
	// prevent subsequent runs of this workflow from re-attempting the same
	// version check, potentially creating a race condition.
	if !restarting && newStatus == app.RunnerStatusActive {
		err = w.checkUpdateNeeded(ctx, heartbeat, healthcheck, runner)
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

// Check if there was a recent restart, and force one if the runner has been running for
// longer than the configured time.
func (w *Workflows) checkRestart(
	ctx workflow.Context,
	heartbeat *app.RunnerHeartBeat,
	healthcheck *app.RunnerHealthCheck,
	runner *app.Runner,
) (bool, error) {
	w.mw.Gauge(ctx, "runner.alivetime", float64(heartbeat.AliveTime.Seconds()), metrics.ToTags(map[string]string{
		"runner_type": string(runner.RunnerGroup.Type),
	})...)

	// TODO(sdboyer) replace with actual value from group settings when actually implementing the call
	ttl := time.Minute * 20
	if heartbeat.AliveTime < time.Second*5 {
		w.mw.Incr(ctx, "runner.restart", metrics.ToTags(map[string]string{
			"runner_type": string(runner.RunnerGroup.Type),
		})...)
	} else if heartbeat.AliveTime > ttl {
		w.mw.Incr(ctx, "runner.ttl_exceeded", metrics.ToTags(map[string]string{
			"runner_type": string(runner.RunnerGroup.Type),
			// "ttl":         runner.RunnerGroup.Settings.TTL.String(),
			"ttl":         ttl.String(),
			"alive_for":   heartbeat.AliveTime.String(),
		})...)

		// TODO(sdboyer) temporary for more granular clarity than the metric gives. Remove once restart is implemented
		l, err := log.WorkflowLogger(ctx)
		if err != nil {
			return false, err
		}
		l.Info("runner ttl exceeded, scheduling restart",
			zap.String("runner_id", runner.ID),
			zap.String("runner_type", string(runner.RunnerGroup.Type)),
			zap.String("ttl", runner.RunnerGroup.Settings.ExpectedVersion),
			zap.String("alive_for", heartbeat.Version),
		)

		w.mw.Event(ctx, &statsd.Event{
			Title: "runner ttl exceeded, scheduling restart",
			Text:  fmt.Sprintf("runner %s has been alive for %s, exceeding TTL of %s", runner.ID, heartbeat.AliveTime, ttl),
			Tags:  metrics.ToTags(map[string]string{
				"runner_type": string(runner.RunnerGroup.Type),
				"runner_id":   runner.ID,
				"ttl":         ttl.String(),
				"alive_for":   heartbeat.AliveTime.String(),
			}),
			AggregationKey: "runner-ttl-restart",
		})

		if generics.SliceContains(runner.ID, testRunners) {
			w.evClient.Send(ctx, runner.ID, &signals.Signal{
				Type:          signals.OperationShutdown,
				HealthCheckID: healthcheck.ID,
			})
		}
		return true, nil
	}

	return false, nil
}

func (w *Workflows) checkUpdateNeeded(
	ctx workflow.Context,
	heartbeat *app.RunnerHeartBeat,
	healthcheck *app.RunnerHealthCheck,
	runner *app.Runner,
) error {
	var needsUpdate bool
	if runner.RunnerGroup.Settings.ExpectedVersion == "latest" {
		needsUpdate = heartbeat.Version != w.cfg.Version
	} else if heartbeat.Version != runner.RunnerGroup.Settings.ExpectedVersion {
		// NOTE(sdboyer) this branch is unreachable until we have a versioning
		// strategy other than latest.
		//
		// However, a lot of older orgs _do_ have something other than `latest`
		// set for their expected version, and as a result they're looping here.
		// So we just never update these, until we have a better strategy.
		needsUpdate = false
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
			zap.String("runner_id", runner.ID),
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
