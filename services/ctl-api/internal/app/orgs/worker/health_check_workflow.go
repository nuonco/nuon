package worker

import (
	"fmt"
	"time"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/worker/activities"
	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

type HealthCheckRequest struct {
	OrgID       string
	SandboxMode bool
}

const (
	// the health check runs each minute
	healthCheckWorkflowCronTab string = "* * * * *"

	// default ping waypoint timeout
	defaultPingWaypointTimeout time.Duration = time.Second * 10
)

func healthCheckWorkflowID(orgID string) string {
	return fmt.Sprintf("%s-health-check", orgID)
}

func (w *Workflows) startHealthCheckWorkflow(ctx workflow.Context, req HealthCheckRequest) {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            healthCheckWorkflowID(req.OrgID),
		CronSchedule:          healthCheckWorkflowCronTab,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, w.OrgHealthCheck, req)
}

func (w *Workflows) OrgHealthCheck(ctx workflow.Context, req HealthCheckRequest) error {
	var healthCheck app.OrgHealthCheck
	if err := w.defaultExecGetActivity(ctx, w.acts.CreateHealthCheck, activities.CreateHealthCheckRequest{
		OrgID: req.OrgID,
	}, &healthCheck); err != nil {
		return fmt.Errorf("unable to create org health check: %w", err)
	}

	var org app.Org
	if err := w.defaultExecGetActivity(ctx, w.acts.Get, activities.GetRequest{
		OrgID: req.OrgID,
	}, &org); err != nil {
		w.updateHealthCheckStatus(ctx, healthCheck.ID, app.OrgHealthCheckStatusError, "unable to get org from database")
		return fmt.Errorf("unable to get org: %w", err)
	}

	if org.Status != string(StatusActive) {
		w.updateHealthCheckStatus(ctx, healthCheck.ID, app.OrgHealthCheckStatus(org.Status), org.StatusDescription)
		return nil
	}

	if req.SandboxMode {
		w.updateHealthCheckStatus(ctx, healthCheck.ID, app.OrgHealthCheckStatusOK, "ok (sandbox mode)")
		return nil
	}

	ctx = workflow.WithStartToCloseTimeout(ctx, defaultPingWaypointTimeout)
	ctx = workflow.WithRetryPolicy(ctx, temporal.RetryPolicy{
		MaximumAttempts: 1,
	})
	if err := w.defaultExecGetActivity(ctx, w.acts.PingWaypointServer, activities.PingWaypointServerRequest{
		OrgID: req.OrgID,
	}, &healthCheck); err != nil {
		w.updateHealthCheckStatus(ctx, healthCheck.ID, app.OrgHealthCheckStatusError, "unable to ping server")
		return nil
	}

	w.updateHealthCheckStatus(ctx, healthCheck.ID, app.OrgHealthCheckStatusOK, "server is active and reachable")
	return nil
}
