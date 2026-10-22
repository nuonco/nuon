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

func (s *Signal) Execute(ctx workflow.Context) error {
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
	if job.DemandEnabled(ctx) {
		return s.runSequence(ctx, install, logStream)
	}

	defer func() {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
	}()

	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	l := workflow.GetLogger(ctx)

	if err := s.updateStepTarget(ctx); err != nil {
		return err
	}

	if s.SandboxMode {
		l.Debug("skipping sync secrets in sandbox mode")
		return nil
	}

	jobID, err := s.prepareSyncSecretsJob(ctx, install)
	if err != nil {
		return err
	}
	if jobID == "" {
		return nil
	}

	l.Info("queueing job and waiting on execution")
	status, err := job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		JobID:    jobID,
		RunnerID: install.RunnerID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("queue-signal-%s-execute-job-%s", install.ID, jobID),
	})
	if err != nil {
		return fmt.Errorf("unable to execute job: %w", err)
	}

	if status != app.RunnerJobStatusFinished {
		l.Error("runner job status was not successful", zap.Any("status", status))
		return fmt.Errorf("unable to sync secrets: %w", err)
	}

	return s.generateState(ctx, jobID)
}

func (s *Signal) runSequence(ctx workflow.Context, install *app.Install, logStream *app.LogStream) error {
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)
	finish := func(ctx workflow.Context, err error) error {
		activities.AwaitCloseLogStreamByLogStreamID(ctx, logStream.ID)
		return err
	}

	if err := s.updateStepTarget(ctx); err != nil {
		return finish(ctx, err)
	}

	var phases []job.Phase
	if s.SandboxMode {
		workflow.GetLogger(ctx).Debug("skipping sync secrets in sandbox mode")
	} else {
		phases = []job.Phase{{
			Key: checkpointSyncSecrets,
			Prepare: func(ctx workflow.Context) (string, error) {
				return s.prepareSyncSecretsJob(ctx, install)
			},
			Finalize: s.generateState,
		}}
	}

	return (job.Sequence{
		Phases: phases,
		Finish: finish,
	}).Run(ctx)
}

func (s *Signal) updateStepTarget(ctx workflow.Context) error {
	if s.WorkflowStepID == "" {
		return nil
	}
	if err := activities.AwaitUpdateInstallWorkflowStepTarget(ctx, activities.UpdateInstallWorkflowStepTargetRequest{
		StepID:         s.WorkflowStepID,
		StepTargetID:   s.WorkflowStepID,
		StepTargetType: "install_workflow_steps",
	}); err != nil {
		return errors.Wrap(err, "unable to update install workflow step target")
	}
	return nil
}

func (s *Signal) prepareSyncSecretsJob(ctx workflow.Context, install *app.Install) (string, error) {
	l := workflow.GetLogger(ctx)

	l.Info("creating plan")
	syncSecretsPlan, err := plan.AwaitCreateSyncSecretsPlan(ctx, &plan.CreateSyncSecretsPlanRequest{
		InstallID: install.ID,
	}, &workflow.ChildWorkflowOptions{
		WorkflowID: fmt.Sprintf("%s-create-sync-secrets-plan", workflow.GetInfo(ctx).WorkflowExecution.ID),
	})
	if err != nil {
		return "", errors.Wrap(err, "unable to create plan")
	}

	if len(syncSecretsPlan.KubernetesSecrets) < 1 {
		l.Debug("no secrets to sync")
		return "", nil
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
		return "", fmt.Errorf("unable to create runner job: %w", err)
	}
	s.runnerJobID = runnerJob.ID

	planJSON, err := json.Marshal(syncSecretsPlan)
	if err != nil {
		return "", errors.Wrap(err, "unable to create json")
	}

	// Deprecated: for now we dual write both the plan json and the composite plan
	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			SyncSecretsPlan: syncSecretsPlan,
		},
	}); err != nil {
		return "", fmt.Errorf("unable to save plan: %w", err)
	}

	return runnerJob.ID, nil
}

func (s *Signal) generateState(ctx workflow.Context, jobID string) error {
	return stategen.HintOrGenerate(ctx, stategen.Request{
		InstallID:       s.InstallID,
		Targets:         statemanager.TargetsForHint(statemanager.HintSecretsUpdated, ""),
		ForceAll:        true,
		TriggeredByID:   jobID,
		TriggeredByType: "runner_jobs",
	})
}
