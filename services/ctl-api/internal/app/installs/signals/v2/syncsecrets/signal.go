package syncsecrets

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
)

const SignalType signal.SignalType = "sync-secrets"

type Signal struct {
	signal.Hooks
	InstallID      string
	WorkflowStepID string
	SandboxMode    bool
}

var _ signal.Signal = &Signal{}
var _ signal.SignalWithStepContext = (*Signal)(nil)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) SetStepContext(stepID, flowID string) {
	s.WorkflowStepID = stepID
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install id is required")
	}

	// Validate install exists
	_, err := activities.AwaitGet(ctx, activities.GetRequest{
		InstallID: s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	install, err := activities.AwaitGet(ctx, activities.GetRequest{
		InstallID: s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		StepID: s.WorkflowStepID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	}()

	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	l := workflow.GetLogger(ctx)

	if s.WorkflowStepID != "" {
		if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
			StepID:         s.WorkflowStepID,
			StepTargetID:   s.WorkflowStepID,
			StepTargetType: "install_workflow_steps",
		}); err != nil {
			return errors.Wrap(err, "unable to update install workflow step target")
		}
	}

	if s.SandboxMode {
		l.Debug("skipping sync secrets in sandbox mode")
		return nil
	}

	l.Info("creating plan")
	syncSecretsPlan, err := plan.AwaitCreateSyncSecretsPlan(ctx, &plan.CreateSyncSecretsPlanRequest{
		InstallID: install.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s-create-sync-secrets-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		return errors.Wrap(err, "unable to create plan")
	}

	if len(syncSecretsPlan.KubernetesSecrets) < 1 {
		l.Debug("no secrets to sync")
		return nil
	}

	// create the job
	l.Info("creating sync secrets job")
	runnerJob, err := activities.AwaitCreateSyncSecretsJob(ctx, &activities.CreateSyncSecretsJobRequest{
		RunnerID:  install.RunnerID,
		InstallID: install.ID,
		OwnerType: "install_workflow_steps",
		OwnerID:   s.WorkflowStepID,
		Op:        app.RunnerJobOperationTypeExec,
		Metadata: map[string]string{
			"install_id": install.ID,
		},
	})
	if err != nil {
		return fmt.Errorf("unable to create runner job: %w", err)
	}

	planJSON, err := json.Marshal(syncSecretsPlan)
	if err != nil {
		return errors.Wrap(err, "unable to create json")
	}

	// Deprecated: for now we dual write both the plan json and the composite plan
	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			SyncSecretsPlan: syncSecretsPlan,
		},
	}); err != nil {
		return fmt.Errorf("unable to save plan: %w", err)
	}

	l.Info("queueing job and waiting on execution")
	status, err := job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		JobID:    runnerJob.ID,
		RunnerID: install.RunnerID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("event-loop-%s-execute-job-%s", install.ID, runnerJob.ID),
	})
	if err != nil {
		return fmt.Errorf("unable to execute job: %w", err)
	}

	if status != app.RunnerJobStatusFinished {
		l.Error("runner job status was not successful", zap.Any("status", status))
		return fmt.Errorf("unable to sync secrets: %w", err)
	}

	return nil
}
