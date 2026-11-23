package registryprune

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"github.com/nuonco/nuon/pkg/config"
	plantypes "github.com/nuonco/nuon/pkg/plans/types"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/activities"
	"github.com/nuonco/nuon/services/ctl-api/internal/app/installs/worker/plan"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/cctx"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/log"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/queue/signal"
	"github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job"
	jobactivities "github.com/nuonco/nuon/services/ctl-api/internal/pkg/workflows/job/activities"
)

const SignalType signal.SignalType = "registry-prune"

const (
	GracePeriod = 30 * 24 * time.Hour
	MaxDeletes  = 200
)

type Signal struct {
	InstallID          string `json:"install_id"`
	DryRun             bool   `json:"dry_run,omitempty"`
	GracePeriodSeconds *int64 `json:"grace_period_seconds,omitempty"`

	runnerJobID string
}

var (
	_ signal.Signal                     = (*Signal)(nil)
	_ signal.SignalWithLifecycleContext = (*Signal)(nil)
	_ signal.SignalWithCancel           = (*Signal)(nil)
)

func (s *Signal) Type() signal.SignalType {
	return SignalType
}

func (s *Signal) LifecycleContext() signal.SignalLifecycleContext {
	return signal.SignalLifecycleContext{
		InstallID: &s.InstallID,
		Operation: string(SignalType),
	}
}

func (s *Signal) Validate(ctx workflow.Context) error {
	if s.InstallID == "" {
		return fmt.Errorf("install_id is required")
	}
	if _, err := activities.AwaitGetByInstallID(ctx, s.InstallID); err != nil {
		return fmt.Errorf("install not found: %w", err)
	}
	return nil
}

func (s *Signal) Cancel(ctx workflow.Context) error {
	cancelCtx, cancel := workflow.NewDisconnectedContext(ctx)
	defer cancel()
	if s.runnerJobID != "" {
		jobactivities.AwaitPkgWorkflowsJobCancelJobByID(cancelCtx, s.runnerJobID)
	}
	return nil
}

func (s *Signal) Execute(ctx workflow.Context) error {
	l, err := log.WorkflowLogger(ctx)
	if err != nil {
		return errors.Wrap(err, "unable to create logger")
	}
	l = l.With(zap.String("install_id", s.InstallID))

	disabled, err := activities.AwaitIsRunnerDisabled(ctx, activities.IsRunnerDisabledRequest{InstallID: s.InstallID})
	if err != nil {
		return fmt.Errorf("unable to check whether the install runner is disabled: %w", err)
	}
	if disabled.Disabled {
		l.Info("install runner is disabled, skipping registry prune")
		return nil
	}

	grace := GracePeriod
	if s.GracePeriodSeconds != nil {
		grace = time.Duration(*s.GracePeriodSeconds) * time.Second
	}

	tags, err := activities.AwaitGetRegistryPruneTags(ctx, &activities.GetRegistryPruneTagsRequest{
		InstallID:      s.InstallID,
		SharedRepo:     true,
		GracePeriod:    grace,
		IncludeWhenOff: s.DryRun,
	})
	if err != nil {
		return fmt.Errorf("unable to compute registry prune tags: %w", err)
	}

	dryRun := s.DryRun || tags.Mode != config.RegistryRetentionModeEnforce
	if tags.Mode == config.RegistryRetentionModeOff && !s.DryRun {
		l.Info("registry retention is off, skipping registry prune")
		return nil
	}
	if len(tags.OwnedTags) == 0 {
		l.Info("install has no synced tags, skipping registry prune")
		return nil
	}

	wfID := workflow.GetInfo(ctx).WorkflowExecution.ID
	planResp, err := plan.AwaitCreatePruneOCIPlan(ctx, &plan.CreatePruneOCIPlanRequest{
		InstallID:        s.InstallID,
		DryRun:           dryRun,
		OwnedTags:        tags.OwnedTags,
		OwnedTagPatterns: tags.OwnedTagPatterns,
		KeepTags:         tags.KeepTags,
		GracePeriod:      grace,
		MaxDeletes:       MaxDeletes,
		WorkflowID:       fmt.Sprintf("%s-create-oci-prune-plan", wfID),
	})
	if err != nil {
		return errors.Wrap(err, "unable to create prune plan")
	}
	if planResp.Plan == nil {
		l.Info("skipping registry prune", zap.String("reason", planResp.SkipReason))
		return nil
	}

	install, err := activities.AwaitGetByInstallID(ctx, s.InstallID)
	if err != nil {
		return errors.Wrap(err, "unable to get install")
	}

	logStream, err := activities.AwaitCreateLogStream(ctx, activities.CreateLogStreamRequest{
		InstallID: s.InstallID,
	})
	if err != nil {
		return errors.Wrap(err, "unable to create log stream")
	}
	defer func() {
		closeCtx, cancel := workflow.NewDisconnectedContext(ctx)
		defer cancel()
		activities.AwaitCloseLogStreamByLogStreamID(closeCtx, logStream.ID)
	}()
	ctx = cctx.SetLogStreamWorkflowContext(ctx, logStream)

	runnerJob, err := activities.AwaitCreateOCIPruneJob(ctx, &activities.CreateOCIPruneJobRequest{
		RunnerID:    install.RunnerID,
		InstallID:   s.InstallID,
		LogStreamID: logStream.ID,
		Metadata: map[string]string{
			"install_id": s.InstallID,
			"dry_run":    fmt.Sprintf("%t", dryRun),
		},
	})
	if err != nil {
		return errors.Wrap(err, "unable to create runner job")
	}
	s.runnerJobID = runnerJob.ID

	planJSON, err := json.Marshal(planResp.Plan)
	if err != nil {
		return errors.Wrap(err, "unable to marshal prune plan")
	}
	if err := activities.AwaitSaveRunnerJobPlan(ctx, &activities.SaveRunnerJobPlanRequest{
		JobID:    runnerJob.ID,
		PlanJSON: string(planJSON),
		CompositePlan: plantypes.CompositePlan{
			PruneOCIPlan: planResp.Plan,
		},
	}); err != nil {
		return errors.Wrap(err, "unable to save runner job plan")
	}

	if _, err := job.AwaitExecuteJob(ctx, &job.ExecuteJobRequest{
		RunnerID:   install.RunnerID,
		JobID:      runnerJob.ID,
		WorkflowID: fmt.Sprintf("%s-execute-job", wfID),
	}); err != nil {
		return fmt.Errorf("registry prune job failed: %w", err)
	}

	finished, err := activities.AwaitGetJobByID(ctx, runnerJob.ID)
	if err != nil {
		return errors.Wrap(err, "unable to get runner job")
	}
	l.Info("registry prune finished",
		zap.String("runner_job_id", runnerJob.ID),
		zap.Bool("dry_run", dryRun),
		zap.Any("result", finished.ParsedOutputs["prune"]),
	)
	return nil
}
