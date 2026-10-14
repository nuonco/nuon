package syncsecrets

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/helpers/stategen"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"

	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	statemanager "github.com/nuonco/nuon/services/ctl-api/internal/pkg/state"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
)

const SignalType signal.SignalType = "sync-secrets"

type Signal struct {
	InstallID      string
	WorkflowStepID string
	SandboxMode    bool

	runnerJobID string
}

var _ signal.Signal = &Signal{}
var _ signal.SignalWithStepContext = (*Signal)(nil)
var _ signal.SignalWithAutoRetry = (*Signal)(nil)
var _ signal.SignalWithMaxRetries = (*Signal)(nil)
var _ signal.SignalWithMaxAutoRetries = (*Signal)(nil)
var _ signal.SignalWithCancel = (*Signal)(nil)
var _ signal.SignalWithTimeout = (*Signal)(nil)

func (s *Signal) AutoRetry() bool                       { return true }
func (s *Signal) MaxRetries() int                       { return 3 }
func (s *Signal) MaxAutoRetries(_ workflow.Context) int { return 3 }
func (s *Signal) Timeout() time.Duration                { return 2 * time.Hour }

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	jobID := s.runnerJobID
	if jobID == "" {
		jobID = activeCheckpointJobID(cancelCtx)
	}
	if jobID != "" {
		jobactivities.AwaitPkgWorkflowsJobCancelJobByID(cancelCtx, jobID)
	}
	return nil
}

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

func (s *Signal) Execute(ctx workflow.Context) (err error) {
	install, err := activities.AwaitGet(ctx, activities.GetRequest{
		InstallID: s.InstallID,
	})
	if err != nil {
		return fmt.Errorf("unable to get install: %w", err)
	}

	var lsCheckpoint logStreamCheckpoint
	if _, err := job.LoadCheckpoint(ctx, checkpointLogStream, &lsCheckpoint); err != nil {
		return errors.Wrap(err, "unable to load log stream checkpoint")
	}

	var logStream *app.LogStream
	if lsCheckpoint.LogStreamID != "" {
		logStream, err = activities.AwaitGetLogStreamByLogStreamID(ctx, lsCheckpoint.LogStreamID)
		if err != nil {
			return errors.Wrap(err, "unable to get log stream")
		}
	} else {
		logStream, err = activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
			StepID: s.WorkflowStepID,
		})
		if err != nil {
			return errors.Wrap(err, "unable to create log stream")
		}
		lsCheckpoint.LogStreamID = logStream.ID
		if err := job.SaveCheckpoint(ctx, checkpointLogStream, lsCheckpoint); err != nil {
			return errors.Wrap(err, "unable to checkpoint log stream")
		}
	}
	defer func() {
		if !job.IsSuspended(err) {
			activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
		}
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

	var cp phaseCheckpoint
	if _, err := job.LoadCheckpoint(ctx, checkpointSyncSecrets, &cp); err != nil {
		return errors.Wrap(err, "unable to load sync secrets checkpoint")
	}
	if cp.Finalized {
		return nil
	}

	if cp.JobID == "" {
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
		s.runnerJobID = runnerJob.ID

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

		cp.JobID = runnerJob.ID
		if err := job.SaveCheckpoint(ctx, checkpointSyncSecrets, cp); err != nil {
			return errors.Wrap(err, "unable to checkpoint sync secrets job")
		}
	}
	s.runnerJobID = cp.JobID

	l.Info("queueing job and waiting on execution")
	status, err := job.Execute(ctx, &job.ExecuteJobRequest{
		JobID:    cp.JobID,
		RunnerID: install.RunnerID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("queue-signal-%s-execute-job-%s", install.ID, cp.JobID),
	})
	if job.IsSuspended(err) {
		return err
	}
	if err != nil {
		return fmt.Errorf("unable to execute job: %w", err)
	}

	if status != app.RunnerJobStatusFinished {
		l.Error("runner job status was not successful", zap.Any("status", status))
		return fmt.Errorf("unable to sync secrets: %w", err)
	}

	if err := stategen.HintOrGenerate(ctx, stategen.Request{
		InstallID:       s.InstallID,
		Targets:         statemanager.TargetsForHint(statemanager.HintSecretsUpdated, ""),
		ForceAll:        true,
		TriggeredByID:   cp.JobID,
		TriggeredByType: "runner_jobs",
	}); err != nil {
		return err
	}

	cp.Finalized = true
	if err := job.SaveCheckpoint(ctx, checkpointSyncSecrets, cp); err != nil {
		return errors.Wrap(err, "unable to checkpoint sync secrets finalization")
	}

	return nil
}
