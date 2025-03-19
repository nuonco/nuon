package worker

import (
	"go.temporal.io/sdk/workflow"

	"github.com/pkg/errors"

	"github.com/powertoolsdev/mono/services/ctl-api/internal/app/general/worker/activities"
	orgssignals "github.com/powertoolsdev/mono/services/ctl-api/internal/app/orgs/signals"
	"github.com/powertoolsdev/mono/services/ctl-api/internal/pkg/log"

	enumsv1 "go.temporal.io/api/enums/v1"
)

const (
	restartOrgRunnersWorkflowCronTab string = "*/1 * * * *"
	restartOrgRunnersWorkflowName    string = "general-restart-org-runners"
)

func (w *Workflows) startRestartOrgRunnersWorkflow(ctx workflow.Context) {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            restartOrgRunnersWorkflowName,
		CronSchedule:          restartOrgRunnersWorkflowCronTab,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, w.RestartOrgRunners)
}

func (w *Workflows) RestartOrgRunners(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	l.Debug("restarting org runners")
	orgs, err := activities.AwaitGetOrgs(ctx, activities.GetOrgsRequest{})
	if err != nil {
		return errors.Wrap(err, "unable to get orgs")
	}

	for _, org := range orgs {
		w.ev.Send(ctx, org.ID, &orgssignals.Signal{
			Type: orgssignals.OperationRestartRunners,
		})
	}

	return nil
}
