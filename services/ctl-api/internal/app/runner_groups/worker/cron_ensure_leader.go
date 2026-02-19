package worker

import (
	"fmt"

	enumsv1 "go.temporal.io/api/enums/v1"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	runnersactivities "github.com/nuonco/nuon/services/ctl-api/internal/app/runners/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
)

const ensureLeaderCronTab = "* * * * *"

func ensureLeaderWorkflowID(groupID string) string {
	return fmt.Sprintf("ensure-leader-%s", groupID)
}

func (w *Workflows) startEnsureLeaderWorkflow(ctx workflow.Context, groupID string) {
	cwo := workflow.ChildWorkflowOptions{
		WorkflowID:            ensureLeaderWorkflowID(groupID),
		CronSchedule:          ensureLeaderCronTab,
		WorkflowIDReusePolicy: enumsv1.WORKFLOW_ID_REUSE_POLICY_TERMINATE_IF_RUNNING,
		ParentClosePolicy:     enumsv1.PARENT_CLOSE_POLICY_TERMINATE,
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	workflow.ExecuteChildWorkflow(ctx, w.EnsureLeader, &EnsureLeaderRequest{
		RunnerGroupID: groupID,
	})
}

type EnsureLeaderRequest struct {
	RunnerGroupID string `validate:"required" json:"runner_group_id"`
}

func (w *Workflows) EnsureLeader(ctx workflow.Context, req *EnsureLeaderRequest) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return err
	}

	resp, err := runnersactivities.AwaitGetGroupLeader(ctx, runnersactivities.GetGroupLeaderRequest{
		RunnerGroupID: req.RunnerGroupID,
	})
	if err != nil {
		l.Warn("unable to check group leader",
			zap.String("runner_group_id", req.RunnerGroupID),
			zap.Error(err),
		)
		return nil
	}

	if resp.LeaderRunnerID != nil {
		return nil
	}

	l.Info("no leader found for runner group, triggering election",
		zap.String("runner_group_id", req.RunnerGroupID),
	)

	result, err := runnersactivities.AwaitElectLeader(ctx, runnersactivities.ElectLeaderRequest{
		RunnerGroupID: req.RunnerGroupID,
	})
	if err != nil {
		l.Error("ensure-leader election failed",
			zap.String("runner_group_id", req.RunnerGroupID),
			zap.Error(err),
		)
		return nil
	}

	if result.NewLeaderID != "" {
		l.Info("ensure-leader elected new leader",
			zap.String("runner_group_id", req.RunnerGroupID),
			zap.String("new_leader_id", result.NewLeaderID),
		)
	}

	return nil
}
