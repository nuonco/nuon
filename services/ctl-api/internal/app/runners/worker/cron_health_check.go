package worker

import (
	"fmt"
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
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
	tags := map[string]string{
		"status": status,
	}
	defer func() {
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

	// ensure the status is created correctly
	newStatus, err := w.getRunnerHeartBeatsStatus(ctx, req.RunnerID)
	if err != nil {
		status = "error_fetching_heart_beats"
		return errors.Wrap(err, "unable to get status from heart beats")
	}

	_, err = activities.AwaitCreateHealthCheck(ctx, activities.CreateHealthCheckRequest{
		RunnerID: req.RunnerID,
		Status:   newStatus,
	})
	if err != nil {
		status = "error"
		return errors.Wrap(err, "unable to create runner health check")
	}

	if newStatus == currentStatus {
		return nil
	}

	// actual status change of a runner
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

func (w *Workflows) getRunnerHeartBeatsStatus(ctx workflow.Context, runnerID string) (app.RunnerStatus, error) {
	// most recent heart beat
	hb, err := activities.AwaitGetMostRecentHeartBeatRequestByRunnerID(ctx, runnerID)
	if err != nil {
		return app.RunnerStatusUnknown, err
	}
	if hb == nil {
		return app.RunnerStatusError, nil
	}

	// update the runner status
	minHeartBeatTS := workflow.Now(ctx).Add(-heartBeatTimeout)
	if hb.CreatedAt.Before(minHeartBeatTS) {
		return app.RunnerStatusError, nil
	}

	return app.RunnerStatusActive, nil
}
