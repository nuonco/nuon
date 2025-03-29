package worker

import (
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
	orgssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"

	enumsv1 "go.temporal.io/api/enums/v1"
)

const (
	restartOrgEventLoopsWorkflowCronTab string = "0 * * * *"
	restartOrgEventLoopsWorkflowName    string = "general-restart-org-event-loops"
)

func (w *Workflows) startRestartOrgEventLoopsWorkflow(ctx workflow.Context) {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            restartOrgEventLoopsWorkflowName,
		CronSchedule:          restartOrgEventLoopsWorkflowCronTab,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, w.RestartOrgEventLoops)
}

func (w *Workflows) RestartOrgEventLoops(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Info("general workflow execution", zap.String("type", "restart-org-event-loops"))

	l.Debug("restarting org event loops")
	orgs, err := activities.AwaitGetOrgs(ctx, activities.GetOrgsRequest{})
	if err != nil {
		return errors.Wrap(err, "unable to get orgs")
	}

	for _, org := range orgs {
		w.ev.Send(ctx, org.ID, &orgssignals.Signal{
			Type: orgssignals.OperationRestart,
		})
	}

	return nil
}
