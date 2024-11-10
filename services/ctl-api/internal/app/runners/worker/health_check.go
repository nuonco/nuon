package worker

import (
	"fmt"
	"strconv"
	"time"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"gorm.io/gorm"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/pkg/generics"
	"github.com/powertoolsdev/mono/pkg/metrics"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/runners/worker/activities"
)

type HealthCheckRequest struct {
	OrgID       string
	RunnerID    string
	SandboxMode bool
	Type        string
}

const (
	// the health check runs every 15 seconds
	healthCheckWorkflowCronTab string = "*/15 * * * *"

	// heart beat timeout
	heartBeatTimeout time.Duration = time.Second * 15
)

func healthCheckWorkflowID(runnerID string) string {
	return fmt.Sprintf("health-check-%s", runnerID)
}

func (w *Workflows) startHealthCheckWorkflow(ctx workflow.Context, req HealthCheckRequest) {
	l := workflow.GetLogger(ctx)
	l.Info("not starting health check workflow")
	return

	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            healthCheckWorkflowID(req.RunnerID),
		CronSchedule:          healthCheckWorkflowCronTab,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, w.HealthCheck, req)
}

func (w *Workflows) HealthCheck(ctx workflow.Context, req HealthCheckRequest) error {
	defaultTags := map[string]string{
		"sandbox_mode": strconv.FormatBool(req.SandboxMode),
		"runner_id":    req.RunnerID,
		"org_id":       req.OrgID,
		"type":         req.Type,
	}
	startTS := workflow.Now(ctx)
	status := "ok"
	op := ""

	defer func() {
		tags := generics.MergeMap(map[string]string{
			"op":     op,
			"status": status,
		}, defaultTags)
		dur := workflow.Now(ctx).Sub(startTS)

		w.mw.Timing(ctx, "health_check.duration", dur, metrics.ToTags(tags)...)
		w.mw.Incr(ctx, "health_check.count", metrics.ToTags(tags)...)
	}()

	// execute health check
	ctx = workflow.WithRetryPolicy(ctx, temporal.RetryPolicy{
		MaximumAttempts: 1,
	})
	if err := w.execHealthCheck(ctx, req.RunnerID); err != nil {
		return errors.Wrap(err, "unable to execute health check")
	}

	return nil
}

func (w *Workflows) getRunnerStatus(ctx workflow.Context, runnerID string) (app.RunnerStatus, error) {
	status, err := activities.AwaitGetRunnerStatusByID(ctx, runnerID)
	if err != nil {
		return app.RunnerStatusUnknown, errors.Wrap(err, "unable to get runner status")
	}

	if status != app.RunnerStatusActive {
		return status, nil
	}

	// most recent heart beat
	hb, err := activities.AwaitGetMostRecentHeartBeatRequestByRunnerID(ctx, runnerID)
	if err != nil {
		return app.RunnerStatusUnknown, err
	}

	// update the runner status
	minHeartBeatTS := workflow.Now(ctx).Add(-heartBeatTimeout)
	if hb.CreatedAt.Before(minHeartBeatTS) {
		return app.RunnerStatusError, nil
	}

	return status, nil
}

func (w *Workflows) execHealthCheck(ctx workflow.Context, runnerID string) error {
	status, err := activities.AwaitGetRunnerStatusByID(ctx, runnerID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner status")
	}

	// most recent heart beat
	hb, err := activities.AwaitGetMostRecentHeartBeatRequestByRunnerID(ctx, runnerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			w.updateStatus(ctx, runnerID, app.RunnerStatusError, "no runner heart beat found")
			return nil
		}

		return err
	}

	// update the runner status
	minHeartBeatTS := workflow.Now(ctx).Add(-heartBeatTimeout)
	newStatus := app.RunnerStatusActive
	newStatusDescription := "active"
	if hb.CreatedAt.Before(minHeartBeatTS) {
		newStatus = app.RunnerStatusError
		newStatusDescription = fmt.Sprintf("no heart beat for %.1f seconds expected less than %.1f", workflow.Now(ctx).Sub(hb.CreatedAt).Seconds(), heartBeatTimeout.Seconds())
	}

	if newStatus == status {
		return nil
	}
	w.updateStatus(ctx, runnerID, newStatus, newStatusDescription)

	return nil
}
